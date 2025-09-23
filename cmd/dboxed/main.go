package main

import (
	"github.com/dboxed/dboxed/cmd/dboxed/cli"

	versionpkg "github.com/dboxed/dboxed/pkg/version"
)

// set via ldflags
var version = ""

func main() {
	// was it set via -ldflags -X
	if //goland:noinspection ALL
	version != "" {
		versionpkg.Version = version
	}

	cli.Execute()
}
