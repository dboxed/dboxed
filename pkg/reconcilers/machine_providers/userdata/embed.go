package userdata

import (
	"bytes"
	"embed"
	"text/template"
)

//go:embed *
var f embed.FS

var templates, _ = template.New("").ParseFS(f, "*")

func GetUserdata(dboxedVersion string, apiUrl string, apiToken string, machineId string) string {
	buf := bytes.NewBuffer(nil)
	err := templates.Lookup("userdata.yaml").Execute(buf, map[string]any{
		"DboxedVersion": dboxedVersion,
		"ApiUrl":        apiUrl,
		"ApiToken":      apiToken,
		"MachineId":     machineId,
	})
	if err != nil {
		panic(err)
	}
	return buf.String()
}
