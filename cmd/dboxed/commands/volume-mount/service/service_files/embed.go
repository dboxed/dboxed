package service_files

import (
	"bytes"
	"embed"
	"text/template"
	"time"
)

//go:embed *
var f embed.FS

var templates, _ = template.New("").ParseFS(f, "*")

type S6RunScripts struct {
	Run    string
	RunLog string
}

func GetS6RunScripts(workDir string, mountName string, volumeUuid string, backupInterval time.Duration) (*S6RunScripts, error) {
	data := map[string]any{
		"DboxedWorkdir":  workDir,
		"MountName":      mountName,
		"VolumeUuid":     volumeUuid,
		"BackupInterval": backupInterval.String(),
	}

	runBuf := bytes.NewBuffer(nil)
	runLogBuf := bytes.NewBuffer(nil)

	err := templates.Lookup("s6-run").Execute(runBuf, data)
	if err != nil {
		panic(err)
	}
	err = templates.Lookup("s6-run-log").Execute(runLogBuf, data)
	if err != nil {
		panic(err)
	}
	return &S6RunScripts{
		Run:    runBuf.String(),
		RunLog: runLogBuf.String(),
	}, nil
}
