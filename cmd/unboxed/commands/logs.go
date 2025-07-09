package commands

import (
	"context"
	"fmt"
	"github.com/koobox/unboxed/cmd/unboxed/flags"
	"github.com/koobox/unboxed/pkg/logs"
	"github.com/koobox/unboxed/pkg/logs/jsonlog"
	"io"
	"os"
	"path/filepath"
)

// same as RFC3339Nano but with trailing zeros
const timeFormat = "2006-01-02T15:04:05.000000000Z07:00"

type LogsCmd struct {
	BoxName       string `help:"Specify the box name" required:"" arg:""`
	ContainerName string `help:"Specify the container name" required:"" arg:""`

	Follow     bool `help:"Follow the logs" short:"f"`
	Tail       int  `help:"Number of lines to show from the end of the log" short:"n" default:"-1"`
	Timestamps bool `help:"Show timestamps" short:"t"`
}

func (cmd *LogsCmd) Run(g *flags.GlobalFlags) error {
	ctx := context.Background()

	logFile := filepath.Join(g.WorkDir, fmt.Sprintf("boxes/%s/containers/%s/logs/logs.json", cmd.BoxName, cmd.ContainerName))
	return logs.TailJsonLogs(ctx, logFile, cmd.Follow, cmd.Tail, func(j *jsonlog.JSONLog) {
		var s io.Writer
		switch j.Stream {
		case "stdout":
			s = os.Stdout
		case "stderr":
			s = os.Stderr
		default:
			s = os.Stderr
		}
		var l string
		if cmd.Timestamps {
			l = fmt.Sprintf("%s %s", j.Created.Format(timeFormat), j.Log)
		} else {
			l = j.Log
		}
		l += "\n"
		_, _ = s.Write([]byte(l))
	})
}
