package volume_backup

import (
	"fmt"
)

func BuildBackupTags(volumeProviderId int64, volumeId *int64, volumeUuid *string, lockId *string) []string {
	tags := []string{
		"dboxed-volume",
		fmt.Sprintf("dboxed-volume-provider-id-%d", volumeProviderId),
	}
	if volumeId != nil {
		tags = append(tags, fmt.Sprintf("dboxed-volume-id-%d", *volumeId))
	}
	if volumeUuid != nil {
		tags = append(tags, fmt.Sprintf("dboxed-volume-uuid-%s", *volumeUuid))
	}
	if lockId != nil {
		tags = append(tags, fmt.Sprintf("dboxed-volume-lock-id-%s", *lockId))
	}
	
	return tags
}
