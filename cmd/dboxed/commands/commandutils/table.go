package commandutils

import (
	"fmt"
	"io"
	"reflect"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/dboxed/dboxed/pkg/util"
)

type TableItem interface {
}

func FormatTable[T TableItem](l []T, showIds bool) string {
	t := reflect.TypeFor[T]()

	// Extract headers and determine which columns to show
	var headers []string
	var columnIndices []int
	for i := range t.NumField() {
		f := t.Field(i)
		col := f.Tag.Get("col")
		idCol := f.Tag.Get("id") == "true"

		// Skip ID columns when showIds is false
		if idCol && !showIds {
			continue
		}

		headers = append(headers, col)
		columnIndices = append(columnIndices, i)
	}

	// Extract rows
	var rows [][]string
	for _, e := range l {
		v := reflect.Indirect(reflect.ValueOf(e))
		var row []string
		for _, i := range columnIndices {
			f := v.Field(i)
			row = append(row, FormatColumn(f.Interface()))
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

func FormatColumn(v any) string {
	switch v.(type) {
	case time.Time:
		return FormatTime(util.Ptr(v.(time.Time)))
	case *time.Time:
		return FormatTime(v.(*time.Time))
	default:
		return fmt.Sprintf("%v", v)
	}
}

func PrintTable[T TableItem](w io.Writer, l []T, showIds bool) error {
	s := FormatTable(l, showIds)
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
