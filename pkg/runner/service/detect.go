//go:build linux

package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/dboxed/dboxed/pkg/util/command_helper"
)

type InitSystem string

const (
	InitSystemSystemd = "systemd"
)

func DetectInitSystem(ctx context.Context) (InitSystem, error) {
	c := command_helper.CommandHelper{
		Command:     "ps",
		Args:        []string{"--no-headers", "-o", "comm", "1"},
		CatchStdout: true,
	}
	err := c.Run(ctx)
	if err != nil {
		return "", err
	}
	s := strings.TrimSpace(string(c.Stdout))
	if s == "systemd" {
		return InitSystemSystemd, nil
	}
	return "", fmt.Errorf("unknown init system")
}
