package volume_backup

import (
	"fmt"
)

func BuildBackupTags(volumeProviderId string, volumeId *string, mountId *string) []string {
	tags := []string{
		"dboxed-volume",
		fmt.Sprintf("dboxed-volume-provider-%s", volumeProviderId),
	}
	if volumeId != nil {
		tags = append(tags, fmt.Sprintf("dboxed-volume-%s", *volumeId))
	}
	if mountId != nil {
		tags = append(tags, fmt.Sprintf("dboxed-volume-mount-%s", *mountId))
	}

	return tags
}
