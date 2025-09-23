package main

import (
	"embed"
	"os"

	"github.com/dboxed/dboxed/pkg/server/db/migration/schematemplates"
)

//go:embed *
var E embed.FS

func main() {
	err := schematemplates.RenderSchemas(E, os.Args[1], os.Args[2])
	if err != nil {
		panic(err)
	}
}
