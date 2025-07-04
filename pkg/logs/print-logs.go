package logs

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

func PrintLogs(ctx context.Context, follow bool, lines int, names []string) error {
	var files []string
	for _, n := range names {
		files = append(files, fmt.Sprintf("%s.log", n))
	}

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
	args = append(args, files...)
	cmd := exec.CommandContext(ctx, "tail", args...)
	cmd.Dir = RootLogDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
