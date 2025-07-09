package logs

import (
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
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
