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

func GetMachineUnit(workspaceId string, machineId string, clientAuthFile string, extraArgs string) string {
	exe, err := os.Executable()
	if err != nil {
		panic(err)
	}

	buf := bytes.NewBuffer(nil)
	err = templates.Lookup("dboxed-machine.service").Execute(buf, map[string]any{
		"ExePath":        exe,
		"WorkspaceId":    workspaceId,
		"MachineId":      machineId,
		"ClientAuthFile": clientAuthFile,
		"ExtraArgs":      extraArgs,
	})
	if err != nil {
		panic(err)
	}
	return buf.String()
}
