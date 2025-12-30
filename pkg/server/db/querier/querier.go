package querier

import (
	"context"
	"database/sql"
	"fmt"
	"maps"
	"reflect"
	"strings"

	"github.com/dboxed/dboxed/pkg/util"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	_ "github.com/lib/pq"
)

type Querier struct {
	Ctx context.Context

	DB *ReadWriteDB
	TX *sqlx.Tx
}

func (q *Querier) selectDriverQuery(query any) string {
	var resolvedQuery string
	if queryStr, ok := query.(string); ok {
		resolvedQuery = queryStr
	} else if m, ok := query.(map[string]string); ok {
		resolvedQuery, ok = m[q.DB.DriverName()]
		if !ok {
			panic("missing query for driver")
		}
	} else {
		panic("invalid query type")
	}
	return resolvedQuery
}

func (q *Querier) bindNamed(query any, arg interface{}) (string, []any, error) {
	resolvedQuery := q.selectDriverQuery(query)
	if arg == nil {
		return resolvedQuery, nil, nil
	}
	if m, ok := arg.(map[string]any); ok {
		var err error
		resolvedQuery, arg, err = q.replacePlaceholders(resolvedQuery, m)
		if err != nil {
			return "", nil, err
		}
	}
	tmp, args, err := sqlx.BindNamed(sqlx.BindType(q.DB.DriverName()), resolvedQuery, arg)
	if err != nil {
		return "", nil, err
	}
	return tmp, args, nil
}

func (q *Querier) replacePlaceholders(query any, m map[string]any) (string, map[string]any, error) {
	resolvedQuery := q.selectDriverQuery(query)
	retMap := m
	var newMap map[string]any
	for k, v := range m {
		if !strings.HasPrefix(k, "@@") {
			continue
		}
		if newMap == nil {
			newMap = maps.Clone(m)
			retMap = newMap
		}
		delete(newMap, k)

		vs, ok := v.(string)
		if !ok {
			return "", nil, fmt.Errorf("value for %s is not a string", k)
		}
		resolvedQuery = strings.ReplaceAll(resolvedQuery, k, vs)
	}
	return resolvedQuery, retMap, nil
}

func (q *Querier) getSqlxQueryer() sqlx.QueryerContext {
	if q.TX != nil {
		return q.TX
	}
	return q.DB
}

func (q *Querier) GetNamed(dest interface{}, query any, arg interface{}) error {
	query2, args, err := q.bindNamed(query, arg)
	if err != nil {
		return err
	}
	err = sqlx.GetContext(q.Ctx, q.getSqlxQueryer(), dest, query2, args...)
	if err != nil {
		if IsSqlNotFoundError(err) {
			t := reflect.TypeOf(dest)
			tableName := GetTableName2(t)
			return &SqlNotFoundError{
				TableName: tableName,
				Err:       err,
			}
		}

		return err
	}

	return nil
}

func (q *Querier) SelectNamed(dest interface{}, query any, arg interface{}) error {
	query2, args, err := q.bindNamed(query, arg)
	if err != nil {
		return err
	}
	return sqlx.SelectContext(q.Ctx, q.getSqlxQueryer(), dest, query2, args...)
}

func (q *Querier) ExecNamed(query any, arg interface{}) (sql.Result, error) {
	query2, args, err := q.bindNamed(query, arg)
	if err != nil {
		return nil, err
	}

	if q.TX != nil {
		return q.TX.ExecContext(q.Ctx, query2, args...)
	} else {
		if IsForbidAutoTx(q.Ctx) {
			return nil, fmt.Errorf("auto-tx not allowed")
		}

		var ret sql.Result
		err = q.DB.Transaction(q.Ctx, func(tx *sqlx.Tx) (bool, error) {
			var err error
			ret, err = tx.ExecContext(q.Ctx, query2, args...)
			if err != nil {
				return false, err
			}
			return true, nil
		})
		if err != nil {
			return nil, err
		}
		return ret, nil
	}
}

func (q *Querier) ExecOneNamed(query string, arg interface{}) error {
	r, err := q.ExecNamed(query, arg)
	if err != nil {
		return err
	}
	ra, err := r.RowsAffected()
	if err != nil {
		return err
	}
	if ra == 0 {
		return sql.ErrNoRows
	}
	if ra != 1 {
		return fmt.Errorf("unexpected rows_affected")
	}
	return nil
}

func Create[T any](q *Querier, v *T) error {
	l := []*T{v}
	return createOrUpdate(q, l, false, "", true)
}
func CreateOrUpdate[T any](q *Querier, v *T, constraint string) error {
	l := []*T{v}
	return createOrUpdate(q, l, true, constraint, true)
}

func CreateMany[T any](q *Querier, l []*T, returning bool) error {
	return createOrUpdate(q, l, false, "", returning)
}

func CreateManyBatches[T any](q *Querier, l []*T, batchSize int, returning bool) error {
	for i := 0; i < len(l); i += batchSize {
		e := i + batchSize
		if e > len(l) {
			e = len(l)
		}
		b := l[i:e]
		err := CreateMany(q, b, returning)
		if err != nil {
			return err
		}
	}
	return nil
}

func createOrUpdate[T any](q *Querier, l []*T, allowUpdate bool, constraint string, returning bool) error {
	t := reflect.TypeFor[T]()
	table := GetTableName2(t)
	fields, _ := GetStructDBFields[T]()

	var createFields []StructDBField
	var createFieldNames []string
	var returningFieldNames []string
	var conflictSets []string
	args := map[string]any{}

	for _, f := range fields {
		if strings.Contains(f.FieldName, ".") {
			continue
		}
		returningFieldNames = append(returningFieldNames, f.FieldName)

		isUuid := f.StructField.Tag.Get("uuid") == "true"
		isChangeSeq := f.FieldName == "change_seq"
		isOmitCreate := f.StructField.Tag.Get("omitCreate") == "true"
		isOmitOnConflictUpdate := f.StructField.Tag.Get("omitOnConflictUpdate") == "true"

		if isOmitCreate {
			continue
		}

		createFields = append(createFields, f)
		createFieldNames = append(createFieldNames, f.FieldName)

		if !isUuid && !isChangeSeq && !isOmitOnConflictUpdate {
			conflictSets = append(conflictSets, fmt.Sprintf("%s = excluded.%s", f.FieldName, f.FieldName))
		}
	}

	var valuesList []string
	for i, v := range l {
		var values []string

		for _, f := range createFields {
			argName := fmt.Sprintf("i%d_%s", i, f.FieldName)

			isUuid := f.StructField.Tag.Get("uuid") == "true"
			isChangeSeq := f.FieldName == "change_seq"

			value := ":" + argName
			if isUuid {
				u, err := uuid.NewV7()
				if err != nil {
					return err
				}
				args[argName] = u.String()
			} else if isChangeSeq {
				value = "nextval('change_tracking_seq')"
			} else {
				fv := GetStructValueByPath(v, f.Path)
				args[argName] = fv.Interface()
			}

			values = append(values, value)
		}

		valuesList = append(valuesList, fmt.Sprintf("(%s)", strings.Join(values, ", ")))
	}

	query := fmt.Sprintf("insert into \"%s\"\n(%s)\nvalues %s",
		table,
		strings.Join(createFieldNames, ", "),
		strings.Join(valuesList, ",\n  "),
	)
	if allowUpdate {
		query += fmt.Sprintf("\non conflict %s do update set %s",
			constraint,
			strings.Join(conflictSets, ", "),
		)
	}
	if returning {
		query += fmt.Sprintf("\nreturning %s", strings.Join(returningFieldNames, ", "))
	}

	var ret []T
	err := q.SelectNamed(&ret, query, args)
	if err != nil {
		return err
	}

	if returning {
		for i, v := range l {
			for _, f := range fields {
				if strings.Contains(f.FieldName, ".") {
					continue
				}
				rv := ret[i]
				fv := GetStructValueByPath(&rv, f.Path)
				tv := GetStructValueByPath(v, f.Path)
				tv.Set(fv)
			}
		}
	}

	return nil
}

func UpdateOneFromStruct[T any](q *Querier, v *T, fields ...string) error {
	dbFields, _ := GetStructDBFields[T]()
	idField, ok := dbFields["id"]
	if !ok {
		return fmt.Errorf("struct has no id field")
	}
	idValue := GetStructValueByPath(v, idField.Path).Interface()
	byFields := map[string]any{
		"id": idValue,
	}

	return UpdateOneByFieldsFromStruct(q, byFields, v, fields...)
}

func UpdateOneByFieldsFromStruct[T any](q *Querier, byFields map[string]any, v *T, fields ...string) error {
	dbFields, _ := GetStructDBFields[T]()
	updateValues := map[string]any{}

	for _, f := range fields {
		sf, ok := dbFields[f]
		if !ok {
			return fmt.Errorf("db field %s not found in struct", f)
		}
		v := GetStructValueByPath(v, sf.Path)
		updateValues[sf.FieldName] = v.Interface()
	}
	return UpdateOneByFields[T](q, byFields, updateValues)
}

func UpdateOneByFields[T any](q *Querier, byFields map[string]any, updateValues map[string]any) error {
	where, args, err := BuildWhere[T](byFields)
	if err != nil {
		return err
	}
	return UpdateOne[T](q, where, args, updateValues)
}

func UpdateOne[T any](q *Querier, where string, whereArgs map[string]any, updateValues map[string]any) error {
	dbFields, _ := GetStructDBFields[T]()

	args := map[string]any{}
	for k, v := range whereArgs {
		args[k] = v
	}

	var sets []string
	for k, v := range updateValues {
		sf, ok := dbFields[k]
		if !ok {
			return fmt.Errorf("db field %s not found in struct", k)
		}

		argName := "_set_" + sf.FieldName
		setValue := fmt.Sprintf(":%s", argName)

		rawSql, ok := v.(RawSqlT)
		if ok {
			setValue = rawSql.SQL
		}

		sets = append(sets, fmt.Sprintf("%s = %s", sf.FieldName, setValue))
		args[argName] = v
	}

	query := fmt.Sprintf("update \"%s\"", GetTableName[T]())
	query += " set " + strings.Join(sets, ", ")
	query += " where " + where

	return q.ExecOneNamed(query, args)
}

func BuildWhere[T any](byFields map[string]any) (string, map[string]any, error) {
	dbFields, _ := GetStructDBFields[T]()

	var where []string
	args := map[string]any{}
	for k, v := range byFields {
		df, ok := dbFields[k]
		if !ok {
			return "", nil, fmt.Errorf("field %s not found", k)
		}

		oin, ok := v.(IsOmitIfNull)
		if ok {
			if !oin.isOmitIfNullValid() {
				continue
			}
		}

		isArg := false
		argName := "_where_" + df.FieldName

		var right string
		if rawSql, ok := v.(RawSqlT); ok {
			right = strings.ReplaceAll(rawSql.SQL, ":", "::")
		} else if util.IsAnyNil(v) {
			right = fmt.Sprintf(" is null")
		} else {
			right = fmt.Sprintf("= :%s", argName)
			isArg = true
		}

		where = append(where, fmt.Sprintf(`%s %s`, df.SelectName, right))
		if isArg {
			args[argName] = v
		}
	}

	whereStr := strings.Join(where, " and ")
	return whereStr, args, nil
}

func BuildSelectWhereQuery[T any](where string, sp *SortAndPage) (string, error) {
	dbFields, dbJoins := GetStructDBFields[T]()

	var selects []string
	for _, f := range dbFields {
		selects = append(selects, fmt.Sprintf(`%s as "%s"`, f.SelectName, f.FieldName))
	}

	var joins []string
	for _, j := range dbJoins {
		joins = append(joins, fmt.Sprintf(`%s join "%s" on "%s"."%s" = "%s"."%s"`,
			j.Type, j.RightTableName, j.LeftTableName, j.LeftIDField, j.RightTableName, j.RightIDField))
	}

	query := fmt.Sprintf("select %s", strings.Join(selects, ",\n  "))
	query += fmt.Sprintf("\nfrom \"%s\"", GetTableName[T]())
	if len(joins) != 0 {
		query += "\n  " + strings.Join(joins, "\n  ")
	}
	if len(where) != 0 {
		query += fmt.Sprintf("\nwhere %s", where)
	}
	if sp != nil {
		if len(sp.Sort) != 0 {
			var orders []string
			for _, s := range sp.Sort {
				f, ok := dbFields[s.Field]
				if !ok {
					return "", fmt.Errorf("sort field %s not found in %s", s.Field, reflect.TypeFor[T]().Name())
				}
				orders = append(orders, fmt.Sprintf("%s %s", f.SelectName, s.Direction))
			}
			query += "\norder by " + strings.Join(orders, ", ")
		}
		if sp.Limit != nil {
			query += fmt.Sprintf("\nlimit %d", *sp.Limit)
		}
		if sp.Offset != 0 {
			query += fmt.Sprintf("\noffset %d", sp.Offset)
		}
	}
	return query, nil
}

func GetOne[T any](q *Querier, byFields map[string]any) (*T, error) {
	where, args, err := BuildWhere[T](byFields)
	if err != nil {
		return nil, err
	}
	return GetOneWhere[T](q, where, args)
}

func GetOneWhere[T any](q *Querier, where string, args map[string]any) (*T, error) {
	query, err := BuildSelectWhereQuery[T](where, nil)
	if err != nil {
		return nil, err
	}

	var ret T
	err = q.GetNamed(&ret, query, args)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

type onlyId[T HasId] struct {
	ID string `db:"id"`
}

func (v *onlyId[T]) GetTableName() string {
	return GetTableName[T]()
}

func CheckOne[T HasId](q *Querier, byFields map[string]any) error {
	where, args, err := BuildWhere[T](byFields)
	if err != nil {
		return err
	}

	query, err := BuildSelectWhereQuery[onlyId[T]](where, nil)
	if err != nil {
		return err
	}

	var ret onlyId[T]
	err = q.GetNamed(&ret, query, args)
	if err != nil {
		return err
	}
	return nil
}

type SortAndPage struct {
	Sort   []SortField
	Offset int64
	Limit  *int64
}

func GetMany[T any](q *Querier, byFields map[string]any, sp *SortAndPage) ([]T, error) {
	return GetManySorted[T](q, byFields, sp)
}

func GetManySorted[T any](q *Querier, byFields map[string]any, sp *SortAndPage) ([]T, error) {
	where, args, err := BuildWhere[T](byFields)
	if err != nil {
		return nil, err
	}
	return GetManyWhere[T](q, where, args, sp)
}

func GetManyWhere[T any](q *Querier, where string, args map[string]any, sp *SortAndPage) ([]T, error) {
	query, err := BuildSelectWhereQuery[T](where, sp)
	if err != nil {
		return nil, err
	}

	var ret []T
	err = q.SelectNamed(&ret, query, args)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func DeleteOneByStruct[T HasId](q *Querier, v T) error {
	return DeleteOneById[T](q, v.GetId())
}

func DeleteOneById[T any](q *Querier, id string) error {
	return DeleteOneByFields[T](q, map[string]any{
		"id": id,
	})
}

func DeleteOneByFields[T any](q *Querier, byFields map[string]any) error {
	where, args, err := BuildWhere[T](byFields)
	if err != nil {
		return err
	}
	return DeleteOneWhere[T](q, where, args)
}

func DeleteOneWhere[T any](q *Querier, where string, args map[string]any) error {
	query := fmt.Sprintf("delete from \"%s\" where %s", GetTableName[T](), where)
	return q.ExecOneNamed(query, args)
}

func DeleteManyByFields[T any](q *Querier, byFields map[string]any) (int, error) {
	where, args, err := BuildWhere[T](byFields)
	if err != nil {
		return 0, err
	}
	return DeleteManyWhere[T](q, where, args)
}

func DeleteManyWhere[T any](q *Querier, where string, args map[string]any) (int, error) {
	query := fmt.Sprintf("delete from \"%s\" where %s", GetTableName[T](), where)
	qr, err := q.ExecNamed(query, args)
	if err != nil {
		return 0, err
	}
	cnt, err := qr.RowsAffected()
	if err != nil {
		return 0, err
	}
	return int(cnt), nil
}

func NewQuerier(ctx context.Context, db *ReadWriteDB, tx *sqlx.Tx) *Querier {
	q := &Querier{
		Ctx: ctx,
		DB:  db,
		TX:  tx,
	}
	return q
}
