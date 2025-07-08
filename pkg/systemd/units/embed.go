package units

import (
	"bytes"
	"embed"
	"text/template"
)

//go:embed *
var f embed.FS

var templates, _ = template.New("").ParseFS(f, "*")

func GetUnboxedUnit(boxName string) string {
	buf := bytes.NewBuffer(nil)
	err := templates.Lookup("unboxed.service").Execute(buf, map[string]any{
		"BoxName": boxName,
	})
	if err != nil {
		panic(err)
	}
	return buf.String()
}
