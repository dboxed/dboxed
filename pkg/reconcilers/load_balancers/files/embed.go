package files

import (
	"bytes"
	"embed"
	"text/template"

	"github.com/Masterminds/sprig/v3"
)

//go:embed *
var f embed.FS

var templates = template.Must(template.New("").Funcs(sprig.FuncMap()).ParseFS(f, "*"))

func GetCaddyComposeFile(caddyVersion string, caddyfile string) (string, error) {
	buf := bytes.NewBuffer(nil)
	err := templates.Lookup("docker-compose.yaml").Execute(buf, map[string]any{
		"CaddyVersion": caddyVersion,
		"Caddyfile":    caddyfile,
	})
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func GetCaddyfile(apiUrl string, apiToken string, workspaeId string, loadBalancerId string) (string, error) {
	buf := bytes.NewBuffer(nil)
	err := templates.Lookup("Caddyfile").Execute(buf, map[string]any{
		"ApiUrl":         apiUrl,
		"ApiToken":       apiToken,
		"WorkspaceId":    workspaeId,
		"LoadBalancerId": loadBalancerId,
	})
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
