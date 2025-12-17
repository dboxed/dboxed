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

	MountId     *string            `json:"mountId,omitempty"`
	MountStatus *VolumeMountStatus `json:"mountStatus,omitempty"`

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

	VolumeTotalSize *int64 `json:"volumeTotalSize,omitempty"`
	VolumeFreeSize  *int64 `json:"volumeFreeSize,omitempty"`

	LastFinishedSnapshotId *string `json:"lastFinishedSnapshotId,omitempty"`

	SnapshotStartTime *time.Time `json:"snapshotStartTime,omitempty"`
	SnapshotEndTime   *time.Time `json:"snapshotEndTime,omitempty"`
}

type VolumeMountStatus struct {
	VolumeId string  `json:"volumeId"`
	MountId  string  `json:"mountId"`
	BoxId    *string `json:"boxId,omitempty"`

	MountTime     time.Time  `json:"mountTime"`
	ReleaseTime   *time.Time `json:"releaseTime,omitempty"`
	ForceReleased bool       `json:"forceReleased"`

	StatusTime time.Time `json:"statusTime"`

	VolumeTotalSize *int64 `json:"volumeTotalSize,omitempty"`
	VolumeFreeSize  *int64 `json:"volumeFreeSize,omitempty"`

	LastFinishedSnapshotId *string `json:"lastFinishedSnapshotId,omitempty"`

	SnapshotStartTime *time.Time `json:"snapshotStartTime,omitempty"`
	SnapshotEndTime   *time.Time `json:"snapshotEndTime,omitempty"`
}

type VolumeReleaseRequest struct {
	MountId string `json:"mountId"`
}

func VolumeFromDB(s dmodel.Volume, attachment *dmodel.BoxVolumeAttachment, volumeProvider *dmodel.VolumeProvider, volumeMountStatus *dmodel.VolumeMountStatus) Volume {
	ret := Volume{
		ID:        s.ID,
		Workspace: s.WorkspaceID,
		CreatedAt: s.CreatedAt,

		Name: s.Name,

		VolumeProviderId:   s.VolumeProviderID,
		VolumeProviderType: s.VolumeProviderType,

		MountId: s.MountId,

		LatestSnapshotId: s.LatestSnapshotId,
	}

	if volumeProvider != nil {
		ret.VolumeProvider = util.Ptr(VolumeProviderFromDB(*volumeProvider))
	}

	if attachment != nil && attachment.VolumeId.Valid {
		ret.Attachment = util.Ptr(VolumeAttachmentFromDB(*attachment, nil, volumeProvider, volumeMountStatus))
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

	if volumeMountStatus != nil && volumeMountStatus.VolumeId.Valid {
		ret.MountStatus = util.Ptr(VolumeMountStatusFromDB(*volumeMountStatus))
	}

	return ret
}

func VolumeMountStatusFromDB(s dmodel.VolumeMountStatus) VolumeMountStatus {
	return VolumeMountStatus{
		VolumeId: s.VolumeId.V,
		MountId:  s.MountId.V,
		BoxId:    s.BoxId,

		MountTime:     s.MountTime.V,
		ReleaseTime:   s.ReleaseTime,
		ForceReleased: s.ForceReleased.V,

		StatusTime: s.StatusTime.V,

		VolumeTotalSize: s.VolumeTotalSize,
		VolumeFreeSize:  s.VolumeFreeSize,

		LastFinishedSnapshotId: s.LastFinishedSnapshotId,

		SnapshotStartTime: s.SnapshotStartTime,
		SnapshotEndTime:   s.SnapshotEndTime,
	}
}

func VolumeAttachmentFromDB(attachment dmodel.BoxVolumeAttachment, volume *dmodel.Volume, volumeProvider *dmodel.VolumeProvider, volumeMountStatus *dmodel.VolumeMountStatus) VolumeAttachment {
	ret := VolumeAttachment{
		BoxID:    attachment.BoxId.V,
		VolumeID: attachment.VolumeId.V,
		RootUid:  attachment.RootUid.V,
		RootGid:  attachment.RootGid.V,
		RootMode: attachment.RootMode.V,
	}

	if volume != nil {
		ret.Volume = util.Ptr(VolumeFromDB(*volume, nil, volumeProvider, volumeMountStatus))
	}

	return ret
}
