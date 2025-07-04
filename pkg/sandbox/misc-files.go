package sandbox

import (
	"fmt"
	"github.com/koobox/unboxed/pkg/types"
	"os"
	"path/filepath"
)

func (rn *Sandbox) writeMiscFiles(c *types.ContainerSpec) error {
	profile := fmt.Sprintf(`export PS1='(unboxed-%s) \u@\h:\w$ '
`, c.Name)
	err := os.WriteFile(filepath.Join(rn.getContainerRoot(c.Name), "etc/profile.d/ps1.sh"), []byte(profile), 0644)
	if err != nil {
		return fmt.Errorf("failed to write ps1.sh: %w", err)
	}

	return nil
}
