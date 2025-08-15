package logs

import (
	"io"

	"gopkg.in/natefinch/lumberjack.v2"
)

func BuildRotatingLogger(logFile string) io.WriteCloser {
	l := &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    100,
		MaxBackups: 3,
		MaxAge:     7,
		Compress:   true,
	}
	return l
}
