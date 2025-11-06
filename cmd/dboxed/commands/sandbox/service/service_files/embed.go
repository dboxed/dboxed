package service_files

import (
	"bytes"
	"embed"
	"os"
	"text/template"
)

//go:embed *
var f embed.FS

var templates, _ = template.New("").ParseFS(f, "*")

func GetDboxedUnit(workspaceId string, boxId string, sandboxName string, clientAuthFile string, extraArgs string) string {
	exe, err := os.Executable()
	if err != nil {
		panic(err)
	}

	buf := bytes.NewBuffer(nil)
	err = templates.Lookup("dboxed-sandbox.service").Execute(buf, map[string]any{
		"ExePath":        exe,
		"SandboxName":    sandboxName,
		"WorkspaceId":    workspaceId,
		"BoxId":          boxId,
		"ClientAuthFile": clientAuthFile,
		"ExtraArgs":      extraArgs,
	})
	if err != nil {
		panic(err)
	}
	return buf.String()
}
