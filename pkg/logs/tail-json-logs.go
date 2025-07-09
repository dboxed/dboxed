package logs

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/koobox/unboxed/pkg/logs/jsonlog"
	"io"
	"os"
	"os/exec"
)

func TailJsonLogs(ctx context.Context, file string, follow bool, lines int, fn func(j *jsonlog.JSONLog)) error {
	var args []string
	if follow {
		args = append(args, "-F")
	}
	if lines == -1 {
		// print all lines
		args = append(args, "-n", "+0")
	} else {
		args = append(args, "-n", fmt.Sprintf("%d", lines))
	}
	args = append(args, file)

	r, w := io.Pipe()

	errCh := make(chan error)
	go func() {
		errCh <- runJsonLineDecode(r, fn)
	}()

	cmd := exec.CommandContext(ctx, "tail", args...)
	cmd.Stdout = w
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}
	_ = w.Close()
	err = <-errCh
	return err
}

func runJsonLineDecode(r io.Reader, fn func(j *jsonlog.JSONLog)) error {
	s := bufio.NewScanner(r)
	s.Buffer(make([]byte, bufio.MaxScanTokenSize), 1024*1024*16)

	for s.Scan() {
		var j jsonlog.JSONLog
		err := json.Unmarshal(s.Bytes(), &j)
		if err != nil {
			continue
		}
		fn(&j)
	}
	return s.Err()
}
