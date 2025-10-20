package commandutils

import (
	"io"
	"reflect"

	"github.com/jedib0t/go-pretty/v6/table"
)

type TableItem interface {
}

func FormatTable[T TableItem](l []T) string {
	t := reflect.TypeFor[T]()

	var header table.Row
	for i := range t.NumField() {
		f := t.Field(i)
		col := f.Tag.Get("col")
		header = append(header, col)
	}

	tw := table.NewWriter()
	tw.AppendHeader(header)
	tw.AppendSeparator()

	for _, e := range l {
		v := reflect.Indirect(reflect.ValueOf(e))
		var row table.Row
		for i := range t.NumField() {
			f := v.Field(i)
			//s, err := fmt.Print(f.Interface())
			//if err != nil {
			//	panic(err)
			//}
			row = append(row, f.Interface())
		}

		tw.AppendRow(row)
	}

	return tw.Render()
}

func PrintTable[T TableItem](w io.Writer, l []T) error {
	s := FormatTable(l)
	_, err := w.Write([]byte(s))
	if err != nil {
		return err
	}
	_, err = w.Write([]byte("\n"))
	if err != nil {
		return err
	}
	return nil
}
