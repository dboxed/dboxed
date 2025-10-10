package main

import (
	"os"

	"github.com/dboxed/dboxed/cmd/dboxed/cli"
	versionpkg "github.com/dboxed/dboxed/pkg/version"
)

// set via ldflags
var version = ""

func main() {
	if len(os.Args) > 1 && os.Args[1] == "init" {
		initLibcontainer()
	}

	// was it set via -ldflags -X
	if //goland:noinspection ALL
	version != "" {
		versionpkg.Version = version
	}

	cli.Execute()
}
