package commandutils

import (
	"fmt"
	"io"
	"reflect"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

type TableItem interface {
}

func FormatTable[T TableItem](l []T) string {
	t := reflect.TypeFor[T]()

	// Extract headers
	var headers []string
	for i := range t.NumField() {
		f := t.Field(i)
		col := f.Tag.Get("col")
		headers = append(headers, col)
	}

	// Extract rows
	var rows [][]string
	for _, e := range l {
		v := reflect.Indirect(reflect.ValueOf(e))
		var row []string
		for i := range t.NumField() {
			f := v.Field(i)
			row = append(row, fmt.Sprintf("%v", f.Interface()))
		}
		rows = append(rows, row)
	}

	// Create lipgloss table with styling
	tw := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("240"))).
		StyleFunc(func(row, col int) lipgloss.Style {
			// Regular cell style
			return lipgloss.NewStyle().
				Foreground(lipgloss.Color("15")).
				Padding(0, 1)
		}).
		Headers(headers...).
		Rows(rows...)

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
