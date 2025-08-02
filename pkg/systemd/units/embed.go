package units

import (
	"bytes"
	"embed"
	"text/template"
)

//go:embed *
var f embed.FS

var templates, _ = template.New("").ParseFS(f, "*")

func GetDboxedUnit(boxName string, extraArgs string) string {
	buf := bytes.NewBuffer(nil)
	err := templates.Lookup("dboxed.service").Execute(buf, map[string]any{
		"BoxName":   boxName,
		"ExtraArgs": extraArgs,
	})
	if err != nil {
		panic(err)
	}
	return buf.String()
}
