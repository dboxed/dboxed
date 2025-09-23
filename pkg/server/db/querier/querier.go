package querier

import (
	"context"
	"database/sql"
	"fmt"
	"maps"
	"reflect"
	"strings"

	"github.com/dboxed/dboxed/pkg/util"
	"github.com/jmoiron/sqlx"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

type Querier struct {
	Ctx context.Context
	DB  *sqlx.DB
	TX  *sqlx.Tx

	E sqlx.ExtContext
}

func (q *Querier) GetDB() *sqlx.DB {
	return q.DB
}

func (q *Querier) selectDriverQuery(query any) string {
	var resolvedQuery string
	if queryStr, ok := query.(string); ok {
		resolvedQuery = queryStr
	} else if m, ok := query.(map[string]string); ok {
		resolvedQuery, ok = m[q.E.DriverName()]
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
	tmp, args, err := sqlx.BindNamed(sqlx.BindType(q.E.DriverName()), resolvedQuery, arg)
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

func (q *Querier) GetNamed(dest interface{}, query any, arg interface{}) error {
	query2, args, err := q.bindNamed(query, arg)
	if err != nil {
		return err
	}
	return sqlx.GetContext(q.Ctx, q.E, dest, query2, args...)
}

func (q *Querier) SelectNamed(dest interface{}, query any, arg interface{}) error {
	query2, args, err := q.bindNamed(query, arg)
	if err != nil {
		return err
	}
	return sqlx.SelectContext(q.Ctx, q.E, dest, query2, args...)
}

func (q *Querier) ExecNamed(query any, arg interface{}) (sql.Result, error) {
	query2, args, err := q.bindNamed(query, arg)
	if err != nil {
		return nil, err
	}
	return q.E.ExecContext(q.Ctx, query2, args...)
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
	return createOrUpdate(q, v, false, "")
}
func CreateOrUpdate[T any](q *Querier, v *T, constraint string) error {
	return createOrUpdate(q, v, true, constraint)
}

func createOrUpdate[T any](q *Querier, v *T, allowUpdate bool, constraint string) error {
	t := reflect.TypeFor[T]()
	table := GetTableName2(t)
	fields, _ := GetStructDBFields[T]()

	var createFieldNames []string
	var returningFieldNames []string
	var argsNames []string
	var sets []string
	var conflictSets []string
	args := map[string]any{}
	for _, f := range fields {
		if strings.Contains(f.FieldName, ".") {
			continue
		}
		returningFieldNames = append(returningFieldNames, f.FieldName)

		if f.StructField.Tag.Get("omitCreate") == "true" {
			continue
		}

		createFieldNames = append(createFieldNames, f.FieldName)
		argsNames = append(argsNames, ":"+f.FieldName)
		sets = append(sets, fmt.Sprintf("%s = :%s", f.FieldName, f.FieldName))
		conflictSets = append(conflictSets, fmt.Sprintf("%s = excluded.%s", f.FieldName, f.FieldName))

		fv := GetStructValueByPath(v, f.Path)
		args[f.FieldName] = fv.Interface()
	}

	query := fmt.Sprintf(`insert into "%s" (%s) values(%s)`,
		table,
		strings.Join(createFieldNames, ", "),
		strings.Join(argsNames, ", "),
	)
	if allowUpdate {
		query += fmt.Sprintf(` on conflict(%s) do update set %s`,
			constraint,
			strings.Join(conflictSets, ", "),
		)
	}
	query += fmt.Sprintf(" returning %s", strings.Join(returningFieldNames, ", "))

	var ret T
	err := q.GetNamed(&ret, query, args)
	if err != nil {
		return err
	}

	for _, f := range fields {
		if strings.Contains(f.FieldName, ".") {
			continue
		}
		fv := GetStructValueByPath(&ret, f.Path)
		tv := GetStructValueByPath(v, f.Path)
		tv.Set(fv)
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

		isNil := util.IsAnyNil(v)
		argName := "_where_" + df.FieldName

		var right string
		if rawSql, ok := v.(RawSqlT); ok {
			right = rawSql.SQL
		} else if isNil {
			right = fmt.Sprintf(" is null")
		} else {
			right = fmt.Sprintf("= :%s", argName)
		}

		where = append(where, fmt.Sprintf(`%s %s`, df.SelectName, right))
		if !isNil {
			args[argName] = v
		}
	}

	whereStr := strings.Join(where, " and ")
	return whereStr, args, nil
}

func BuildSelectWhereQuery[T any](where string) (string, error) {
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
	query, err := BuildSelectWhereQuery[T](where)
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

func GetMany[T any](q *Querier, byFields map[string]any) ([]T, error) {
	where, args, err := BuildWhere[T](byFields)
	if err != nil {
		return nil, err
	}
	return GetManyWhere[T](q, where, args)
}

func GetManyWhere[T any](q *Querier, where string, args map[string]any) ([]T, error) {
	query, err := BuildSelectWhereQuery[T](where)
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

func DeleteOneById[T any](q *Querier, id int64) error {
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

func NewQuerier(ctx context.Context, db *sqlx.DB, tx *sqlx.Tx) *Querier {
	var e sqlx.ExtContext
	if tx != nil {
		e = tx
	} else {
		e = db
	}
	q := &Querier{
		Ctx: ctx,
		DB:  db,
		TX:  tx,
		E:   e,
	}
	return q
}

func GetTableName[T any](_ ...T) string {
	t := reflect.TypeFor[T]()
	return GetTableName2(t)
}

func GetTableName2(t reflect.Type) string {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	if reflect.PointerTo(t).Implements(reflect.TypeFor[HasTableName]()) {
		i := reflect.New(t).Interface().(HasTableName)
		return i.GetTableName()
	}

	return util.ToSnakeCase(t.Name())
}
