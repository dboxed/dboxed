package main

import (
	"os"

	dboxed_volume_cli "github.com/dboxed/dboxed-volume/cmd/dboxed-volume/cli"
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

	if os.Args[0] == "dboxed-volume" {
		dboxed_volume_cli.Execute()
	} else {
		cli.Execute()
	}
}
