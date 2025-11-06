package volume_backup

import (
	"fmt"
)

func BuildBackupTags(volumeProviderId string, volumeId *string, lockId *string) []string {
	tags := []string{
		"dboxed-volume",
		fmt.Sprintf("dboxed-volume-provider-id-%s", volumeProviderId),
	}
	if volumeId != nil {
		tags = append(tags, fmt.Sprintf("dboxed-volume-id-%s", *volumeId))
	}
	if lockId != nil {
		tags = append(tags, fmt.Sprintf("dboxed-volume-lock-id-%s", *lockId))
	}

	return tags
}
