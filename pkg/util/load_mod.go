package util

import (
	"context"
	"os/exec"
)

func LoadMod(ctx context.Context, name string) {
	// this is a hack that works inside containers. It loads a module via the Kernel's auto-load feature
	c := exec.CommandContext(ctx, "ip", "link", "show", name)
	_ = c.Run()
}
