package logs

import (
	"fmt"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
	"path/filepath"
)

const RootLogDir = "/var/log/unboxed"

func BuildRotatingLogger(name string) (io.WriteCloser, error) {
	err := os.MkdirAll(RootLogDir, 0700)
	if err != nil {
		return nil, err
	}

	logFile := filepath.Join(RootLogDir, fmt.Sprintf("%s.log", name))
	l := &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    100,
		MaxBackups: 3,
		MaxAge:     7,
		Compress:   true,
	}
	return l, nil
}
