package units

import (
	"bytes"
	"embed"
	"text/template"
)

//go:embed *
var f embed.FS

var templates, _ = template.New("").ParseFS(f, "*")

func GetDboxedUnit(localName string, clientAuthFile string, extraArgs string) string {
	buf := bytes.NewBuffer(nil)
	err := templates.Lookup("dboxed.service").Execute(buf, map[string]any{
		"LocalName":      localName,
		"ClientAuthFile": clientAuthFile,
		"ExtraArgs":      extraArgs,
	})
	if err != nil {
		panic(err)
	}
	return buf.String()
}
