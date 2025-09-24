package models

import (
	"time"

	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/util"
)

type Volume struct {
	ID        int64     `json:"id"`
	Workspace int64     `json:"workspace"`
	CreatedAt time.Time `json:"created_at"`

	Uuid string `json:"uuid"`
	Name string `json:"name"`

	VolumeProvider     int64                     `json:"volume_provider"`
	VolumeProviderType dmodel.VolumeProviderType `json:"volume_provider_type"`

	LockId   *string    `db:"lock_id"`
	LockTime *time.Time `db:"lock_time"`

	LatestSnapshotId *int64 `json:"latest_snapshot_id"`

	Attachment *VolumeAttachment `json:"attachment,omitempty"`

	Rustic *VolumeRustic `json:"rustic"`
}

type VolumeRustic struct {
	FsSize int64  `json:"fs_size"`
	FsType string `json:"fs_type"`
}

type CreateVolume struct {
	Name string `json:"name"`

	VolumeProvider int64 `json:"volume_provider"`

	Rustic *CreateVolumeRustic `json:"rustic,omitempty"`
}

type CreateVolumeRustic struct {
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

	Volume *Volume `json:"volume,omitempty"`
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

type VolumeLockRequest struct {
	PrevLockId *string `json:"prevLockId"`
}

type VolumeReleaseRequest struct {
	LockId string `json:"lockI"`
}

func VolumeFromDB(s dmodel.Volume, attachment *dmodel.BoxVolumeAttachment) Volume {
	ret := Volume{
		ID:        s.ID,
		Workspace: s.WorkspaceID,
		CreatedAt: s.CreatedAt,

		Uuid: s.Uuid,
		Name: s.Name,

		VolumeProvider:     s.VolumeProviderID,
		VolumeProviderType: dmodel.VolumeProviderType(s.VolumeProviderType),

		LockId:   s.LockId,
		LockTime: s.LockTime,

		LatestSnapshotId: s.LatestSnapshotId,
	}

	if attachment != nil {
		ret.Attachment = util.Ptr(VolumeAttachmentFromDB(*attachment, nil))
	}

	if s.Rustic != nil {
		ret.Rustic = &VolumeRustic{
			FsSize: s.Rustic.FsSize.V,
			FsType: s.Rustic.FsType.V,
		}
	}

	return ret
}

func VolumeAttachmentFromDB(attachment dmodel.BoxVolumeAttachment, volume *dmodel.Volume) VolumeAttachment {
	ret := VolumeAttachment{
		BoxID:    attachment.BoxId.V,
		VolumeID: attachment.VolumeId.V,
		RootUid:  attachment.RootUid.V,
		RootGid:  attachment.RootGid.V,
		RootMode: attachment.RootMode.V,
	}

	if volume != nil {
		ret.Volume = util.Ptr(VolumeFromDB(*volume, nil))
	}

	return ret
}
