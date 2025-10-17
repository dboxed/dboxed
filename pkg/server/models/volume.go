package models

import (
	"time"

	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/util"
)

type Volume struct {
	ID        int64     `json:"id"`
	Workspace int64     `json:"workspace"`
	CreatedAt time.Time `json:"createdAt"`

	Uuid string `json:"uuid"`
	Name string `json:"name"`

	VolumeProvider     int64                     `json:"volumeProvider"`
	VolumeProviderType dmodel.VolumeProviderType `json:"volumeProviderType"`

	LockId    *string    `json:"lockId,omitempty"`
	LockTime  *time.Time `json:"lockTime,omitempty"`
	LockBoxId *int64     `json:"lockBoxId,omitempty"`

	LatestSnapshotId *int64 `json:"latestSnapshotId,omitempty"`

	Attachment *VolumeAttachment `json:"attachment,omitempty"`

	Rustic *VolumeRustic `json:"rustic,omitempty"`
}

type VolumeRustic struct {
	Password string `json:"password"`

	FsSize int64  `json:"fsSize"`
	FsType string `json:"fsType"`
}

type CreateVolume struct {
	Name string `json:"name"`

	VolumeProvider int64 `json:"volumeProvider"`

	Rustic *CreateVolumeRustic `json:"rustic,omitempty"`
}

type CreateVolumeRustic struct {
	FsSize int64  `json:"fsSize"`
	FsType string `json:"fsType"`
}

type UpdateVolume struct {
}

type VolumeAttachment struct {
	BoxID    int64 `json:"boxId"`
	VolumeID int64 `json:"volumeId"`

	RootUid  int64  `json:"rootUid"`
	RootGid  int64  `json:"rootGid"`
	RootMode string `json:"rootMode"`

	Volume *Volume `json:"volume,omitempty"`
}

type AttachVolumeRequest struct {
	VolumeId int64 `json:"volumeId"`

	RootUid  int64  `json:"rootUid"`
	RootGid  int64  `json:"rootGid"`
	RootMode string `json:"rootMode"`
}

type UpdateVolumeAttachmentRequest struct {
	RootUid  *int64  `json:"rootUid,omitempty"`
	RootGid  *int64  `json:"rootGid,omitempty"`
	RootMode *string `json:"rootMode,omitempty"`
}

type VolumeLockRequest struct {
	PrevLockId *string `json:"prevLockId,omitempty"`
	BoxId      *int64  `json:"boxId,omitempty"`
}

type VolumeReleaseRequest struct {
	LockId string `json:"lockId"`
}

func VolumeFromDB(s dmodel.Volume, attachment *dmodel.BoxVolumeAttachment, volumeProvider *dmodel.VolumeProvider) Volume {
	ret := Volume{
		ID:        s.ID,
		Workspace: s.WorkspaceID,
		CreatedAt: s.CreatedAt,

		Uuid: s.Uuid,
		Name: s.Name,

		VolumeProvider:     s.VolumeProviderID,
		VolumeProviderType: s.VolumeProviderType,

		LockId:    s.LockId,
		LockTime:  s.LockTime,
		LockBoxId: s.LockBoxId,

		LatestSnapshotId: s.LatestSnapshotId,
	}

	if attachment != nil && attachment.VolumeId.Valid {
		ret.Attachment = util.Ptr(VolumeAttachmentFromDB(*attachment, nil, volumeProvider))
	}

	if s.Rustic != nil && s.Rustic.ID.Valid {
		ret.Rustic = &VolumeRustic{
			FsSize: s.Rustic.FsSize.V,
			FsType: s.Rustic.FsType.V,
		}
		if volumeProvider != nil {
			ret.Rustic.Password = volumeProvider.Rustic.Password.V
		}
	}

	return ret
}

func VolumeAttachmentFromDB(attachment dmodel.BoxVolumeAttachment, volume *dmodel.Volume, volumeProvider *dmodel.VolumeProvider) VolumeAttachment {
	ret := VolumeAttachment{
		BoxID:    attachment.BoxId.V,
		VolumeID: attachment.VolumeId.V,
		RootUid:  attachment.RootUid.V,
		RootGid:  attachment.RootGid.V,
		RootMode: attachment.RootMode.V,
	}

	if volume != nil {
		ret.Volume = util.Ptr(VolumeFromDB(*volume, nil, volumeProvider))
	}

	return ret
}
