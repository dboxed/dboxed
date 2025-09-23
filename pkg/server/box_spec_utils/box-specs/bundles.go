package box_specs

import (
	"fmt"

	"github.com/dboxed/dboxed/pkg/boxspec"
)

func addFileToBundle(b *boxspec.FileBundle, path string, data string, mode uint32) {
	b.Files = append(b.Files, boxspec.FileBundleEntry{
		Path:       path,
		StringData: data,
		Mode:       fmt.Sprintf("%03o", mode),
	})
}
