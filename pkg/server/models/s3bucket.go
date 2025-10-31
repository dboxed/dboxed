package models

import (
	"time"

	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
)

type S3Bucket struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	Workspace int64     `json:"workspace"`

	Status        string `json:"status"`
	StatusDetails string `json:"statusDetails"`

	Endpoint string `json:"endpoint"`
	Bucket   string `json:"bucket"`
}

type CreateS3Bucket struct {
	Endpoint        string `json:"endpoint"`
	Bucket          string `json:"bucket"`
	AccessKeyId     string `json:"accessKeyId"`
	SecretAccessKey string `json:"secretAccessKey"`
}

type UpdateS3Bucket struct {
	Endpoint        *string `json:"endpoint,omitempty"`
	Bucket          *string `json:"bucket,omitempty"`
	AccessKeyId     *string `json:"accessKeyId,omitempty"`
	SecretAccessKey *string `json:"secretAccessKey,omitempty"`
}

func S3BucketFromDB(v dmodel.S3Bucket) S3Bucket {
	ret := S3Bucket{
		ID:            v.ID,
		CreatedAt:     v.CreatedAt,
		Workspace:     v.WorkspaceID,
		Status:        v.ReconcileStatus.ReconcileStatus.V,
		StatusDetails: v.ReconcileStatus.ReconcileStatusDetails.V,

		Endpoint: v.Endpoint,
		Bucket:   v.Bucket,
	}
	return ret
}
