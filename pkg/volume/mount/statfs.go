package mount

import (
	"fmt"

	"golang.org/x/sys/unix"
)

type FileSystemStats struct {
	TotalSize int64
	FreeSize  int64
}

func StatFS(mountDir string) (*FileSystemStats, error) {
	var st unix.Statfs_t
	err := unix.Statfs(mountDir, &st)
	if err != nil {
		return nil, fmt.Errorf("failed to determine filesystem stats: %w", err)
	}

	ret := FileSystemStats{
		TotalSize: int64(int64(st.Bsize) * int64(st.Blocks)),
		FreeSize:  int64(int64(st.Bsize) * int64(st.Bfree)),
	}

	return &ret, nil
}
