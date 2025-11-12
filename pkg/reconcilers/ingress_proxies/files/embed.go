package files

import (
	"bytes"
	"embed"
	"text/template"
)

//go:embed *
var f embed.FS

var templates, _ = template.New("").ParseFS(f, "*")

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
