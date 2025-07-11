package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/koobox/unboxed/cmd/unboxed/flags"
	"github.com/koobox/unboxed/pkg/logs"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
)

type InitWrapperCmd struct {
	Args []string `arg:"" optional:"" passthrough:""`

	logRot io.WriteCloser
	stdout io.WriteCloser
	stderr io.WriteCloser
	log    io.WriteCloser

	logPrefix   string
	logToStderr bool
}

func (cmd *InitWrapperCmd) Run(g *flags.GlobalFlags) error {
	exitCode := cmd.run2(g)
	cmd.writeLog("exiting init-wrapper")
	os.Exit(exitCode)
	return nil
}

func (cmd *InitWrapperCmd) initLogging() bool {
	logFile := os.Getenv("UNBOXED_LOG_FILE")
	logToStderrStr := os.Getenv("UNBOXED_LOG_STDERR")
	cmd.logPrefix = os.Getenv("UNBOXED_LOG_PREFIX")
	_ = os.Unsetenv("UNBOXED_LOG_FILE")
	_ = os.Unsetenv("UNBOXED_LOG_STDERR")
	_ = os.Unsetenv("UNBOXED_LOG_PREFIX")

	cmd.logToStderr, _ = strconv.ParseBool(logToStderrStr)

	if logFile == "" {
		cmd.writeStderr("missing UNBOXED_LOG_FILE")
		return false
	}
	err := os.WriteFile(logFile+".test", nil, 0600)
	if err != nil {
		cmd.writeStderr("can't write to log directory: %s", err.Error())
		return false
	}
	err = os.Remove(logFile + ".test")
	if err != nil {
		cmd.writeStderr("can't delete test file: %s", err.Error())
		return false
	}

	cmd.logRot = logs.BuildRotatingLogger(logFile)
	cmd.stdout = logs.NewJsonFileLogger(cmd.logRot, "stdout")
	cmd.stderr = logs.NewJsonFileLogger(cmd.logRot, "stderr")
	cmd.log = logs.NewJsonFileLogger(cmd.logRot, "init-wrapper")

	return true
}

func (cmd *InitWrapperCmd) closeLogging() {
	cmd.writeLog("close logging")
	_ = cmd.log.Close()
	_ = cmd.stderr.Close()
	_ = cmd.stdout.Close()
	_ = cmd.logRot.Close()
	cmd.log = nil
	cmd.writeLog("logs closed")
}

func (cmd *InitWrapperCmd) formatLogLine(f string, args ...any) string {
	l := fmt.Sprintf(f+"\n", args...)
	if cmd.logPrefix != "" {
		l = cmd.logPrefix + ": " + l
	}
	return l
}

func (cmd *InitWrapperCmd) writeStderr(f string, args ...any) {
	_, _ = fmt.Fprint(os.Stderr, cmd.formatLogLine(f, args...))
}

func (cmd *InitWrapperCmd) writeLog(f string, args ...any) {
	l := cmd.formatLogLine(f, args...)
	if cmd.log != nil {
		_, _ = cmd.log.Write([]byte(l))
	}
	if cmd.logToStderr {
		_, _ = os.Stderr.Write([]byte(l))
	}
}

func (cmd *InitWrapperCmd) waitPidLoop() {
	for {
		if wpid, _ := syscall.Wait4(-1, nil, syscall.WNOHANG, nil); wpid <= 0 {
			break
		} else {
			cmd.writeLog("reaped zombie process: pid=%d", wpid)
		}
	}
}

func (cmd *InitWrapperCmd) startReaper() chan os.Signal {
	cmd.writeLog("starting child reaper")

	cmd.waitPidLoop()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGCHLD)
	go func() {
		for s := range sigs {
			if s == syscall.SIGCHLD {
				cmd.writeLog("received %s, try reaping", s.String())
				cmd.waitPidLoop()
			}
		}
		cmd.writeLog("reaper stopped")
	}()
	return sigs
}

func (cmd *InitWrapperCmd) startSignalForward(process *os.Process) chan os.Signal {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGUSR1, syscall.SIGUSR2, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		for s := range sigs {
			cmd.writeLog("signalling process %d: %s", process.Pid, s.String())
			err := process.Signal(s)
			if err != nil {
				cmd.writeLog("failed signalling process: %s", err.Error())
			}
		}
		cmd.writeLog("signal forwarder stopped")
	}()
	return sigs
}

func (cmd *InitWrapperCmd) run2(g *flags.GlobalFlags) int {
	if !cmd.initLogging() {
		return -1
	}
	defer cmd.closeLogging()

	if os.Getpid() == 1 {
		sc := cmd.startReaper()
		defer func() {
			cmd.writeLog("stopping reaper")
			signal.Stop(sc)
			close(sc)
		}()
	}

	j, _ := json.Marshal(cmd.Args)
	cmd.writeLog("starting process: " + string(j))

	c := exec.Command(cmd.Args[0], cmd.Args[1:]...)
	c.Stdout = cmd.stdout
	c.Stderr = cmd.stderr
	c.Stdin = os.Stdin

	err := c.Start()
	if err != nil {
		cmd.writeLog("start returned error: %s", err.Error())
		return -1
	}
	sc := cmd.startSignalForward(c.Process)
	defer func() {
		cmd.writeLog("stopping signal forwarder")
		signal.Stop(sc)
		close(sc)
	}()

	cmd.writeLog("waiting for process exit (pid=%d)", c.Process.Pid)
	err = c.Wait()
	if err != nil {
		cmd.writeLog("wait returned error: %s", err.Error())
		var err2 *exec.ExitError
		if errors.As(err, &err2) {
			return err2.ExitCode()
		} else {
			return -1
		}
	} else {
		cmd.writeLog("wait returned with no error")
		return 0
	}
}
