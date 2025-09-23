package models

import (
	"time"

	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/global"
)

type Volume struct {
	ID        int64     `json:"id"`
	Workspace int64     `json:"workspace"`
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`

	VolumeProvider     int64                     `json:"volume_provider"`
	VolumeProviderType global.VolumeProviderType `json:"volume_provider_type"`

	AttachedToBox *int64 `json:"attached_to_box,omitempty"`

	Dboxed *VolumeDboxed `json:"dboxed"`
}

type VolumeDboxed struct {
	FsSize int64  `json:"fs_size"`
	FsType string `json:"fs_type,omitempty"`
}

type CreateVolume struct {
	Name string `json:"name"`

	VolumeProvider int64 `json:"volume_provider"`

	Dboxed *CreateVolumeDboxed `json:"dboxed,omitempty"`
}

type CreateVolumeDboxed struct {
	FsSize int64  `json:"fs_size"`
	FsType string `json:"fs_type,omitempty"`
}

type UpdateVolume struct {
}

type VolumeAttachment struct {
	BoxID    int64 `json:"box_id"`
	VolumeID int64 `json:"volume_id"`

	RootUid  int64  `json:"root_uid"`
	RootGid  int64  `json:"root_gid"`
	RootMode string `json:"root_mode"`

	Volume Volume `json:"volume"`
}

type AttachVolumeRequest struct {
	VolumeId int64 `json:"volume_id"`

	RootUid  int64  `json:"root_uid"`
	RootGid  int64  `json:"root_gid"`
	RootMode string `json:"root_mode"`
}

type UpdateVolumeAttachmentRequest struct {
	RootUid  *int64  `json:"root_uid,omitempty"`
	RootGid  *int64  `json:"root_gid,omitempty"`
	RootMode *string `json:"root_mode,omitempty"`
}

func VolumeFromDB(s dmodel.Volume, attachment *dmodel.BoxVolumeAttachment) Volume {
	ret := Volume{
		ID:        s.ID,
		Workspace: s.WorkspaceID,
		CreatedAt: s.CreatedAt,
		Name:      s.Name,

		VolumeProvider:     s.VolumeProviderID,
		VolumeProviderType: global.VolumeProviderType(s.VolumeProviderType),
	}

	if attachment != nil {
		ret.AttachedToBox = &attachment.BoxId.V
	}

	if s.Dboxed != nil {
		ret.Dboxed = &VolumeDboxed{
			FsSize: s.Dboxed.FsSize.V,
			FsType: s.Dboxed.FsType.V,
		}
	}

	return ret
}

func VolumeAttachmentFromDB(v dmodel.BoxVolumeAttachmentWithVolume) VolumeAttachment {
	return VolumeAttachment{
		BoxID:    v.BoxId.V,
		VolumeID: v.VolumeId.V,
		RootUid:  v.RootUid.V,
		RootGid:  v.RootGid.V,
		RootMode: v.RootMode.V,
		Volume:   VolumeFromDB(v.Volume, &v.BoxVolumeAttachment),
	}
}
