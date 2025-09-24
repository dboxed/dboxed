package schematemplates

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

var replacements = map[string]map[string]string{
	"postgres": {
		"TYPES_INT_PRIMARY_KEY": "bigserial not null primary key",
		"TYPES_DATETIME":        "timestamptz",
	},
	"sqlite": {
		"TYPES_INT_PRIMARY_KEY": "integer not null primary key autoincrement",
		"TYPES_DATETIME":        "datetime",
	},
}

func renderSchemas(sourceFs fs.FS, dbType string) (map[string]string, error) {
	m := map[string]string{}

	files, err := fs.ReadDir(sourceFs, ".")
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".sql") {
			continue
		}
		b, err := fs.ReadFile(sourceFs, f.Name())
		if err != nil {
			return nil, err
		}

		t, err := template.New("").Parse(string(b))
		if err != nil {
			return nil, err
		}
		buf := bytes.NewBuffer(nil)
		err = t.Execute(buf, map[string]any{
			"DbType": dbType,
		})
		if err != nil {
			return nil, err
		}

		s := buf.String()
		for k, v := range replacements[dbType] {
			s = strings.ReplaceAll(s, k, v)
		}
		m[f.Name()] = s
	}

	return m, nil
}

func RenderSchemas(sourceFs fs.FS, targetDir string, dbType string) error {
	err := os.MkdirAll(targetDir, 0755)
	if err != nil {
		return err
	}

	m, err := renderSchemas(sourceFs, dbType)
	if err != nil {
		return err
	}
	for k, v := range m {
		targetFile := filepath.Join(targetDir, k)
		err = os.WriteFile(targetFile, []byte(v), 0644)
		if err != nil {
			return err
		}
	}

	return nil
}
