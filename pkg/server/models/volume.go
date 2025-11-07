package models

import (
	"time"

	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/util"
)

type Volume struct {
	ID        string    `json:"id"`
	Workspace string    `json:"workspace"`
	CreatedAt time.Time `json:"createdAt"`

	Name string `json:"name"`

	VolumeProviderId   string                    `json:"volumeProviderId"`
	VolumeProviderType dmodel.VolumeProviderType `json:"volumeProviderType"`
	VolumeProvider     *VolumeProvider           `json:"volumeProvider"`

	MountId    *string    `json:"mountId,omitempty"`
	MountTime  *time.Time `json:"mountTime,omitempty"`
	MountBoxId *string    `json:"mountBoxId,omitempty"`

	LatestSnapshotId *string `json:"latestSnapshotId,omitempty"`

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

	VolumeProvider string `json:"volumeProvider"`

	Rustic *CreateVolumeRustic `json:"rustic,omitempty"`
}

type CreateVolumeRustic struct {
	FsSize int64  `json:"fsSize"`
	FsType string `json:"fsType"`
}

type UpdateVolume struct {
}

type VolumeAttachment struct {
	BoxID    string `json:"boxId"`
	VolumeID string `json:"volumeId"`

	RootUid  int64  `json:"rootUid"`
	RootGid  int64  `json:"rootGid"`
	RootMode string `json:"rootMode"`

	Volume *Volume `json:"volume,omitempty"`
}

type AttachVolumeRequest struct {
	VolumeId string `json:"volumeId"`

	RootUid  *int64  `json:"rootUid,omitempty"`
	RootGid  *int64  `json:"rootGid,omitempty"`
	RootMode *string `json:"rootMode,omitempty"`
}

type UpdateVolumeAttachmentRequest struct {
	RootUid  *int64  `json:"rootUid,omitempty"`
	RootGid  *int64  `json:"rootGid,omitempty"`
	RootMode *string `json:"rootMode,omitempty"`
}

type VolumeMountRequest struct {
	BoxId *string `json:"boxId,omitempty"`
}

type VolumeRefreshMountRequest struct {
	MountId string `json:"mountId"`
}

type VolumeReleaseRequest struct {
	MountId string `json:"mountId"`
}

func VolumeFromDB(s dmodel.Volume, attachment *dmodel.BoxVolumeAttachment, volumeProvider *dmodel.VolumeProvider) Volume {
	ret := Volume{
		ID:        s.ID,
		Workspace: s.WorkspaceID,
		CreatedAt: s.CreatedAt,

		Name: s.Name,

		VolumeProviderId:   s.VolumeProviderID,
		VolumeProviderType: s.VolumeProviderType,

		MountId:    s.MountId,
		MountTime:  s.MountTime,
		MountBoxId: s.MountBoxId,

		LatestSnapshotId: s.LatestSnapshotId,
	}

	if volumeProvider != nil {
		ret.VolumeProvider = util.Ptr(VolumeProviderFromDB(*volumeProvider))
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
