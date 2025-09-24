package models

import (
	"time"
)

type S3ProxyPresignPutRequest struct {
	Key string `json:"key"`
}

type S3ProxyPresignPutResult struct {
	PresignedUrl string    `json:"presignedUrl"`
	Expires      time.Time `json:"expires"`
}

type S3ProxyListObjectsRequest struct {
	Prefix string `json:"prefix"`
}

type S3ProxyListObjectsResult struct {
	Objects []S3ObjectInfo `json:"objects"`
}

type S3ProxyRenameObjectRequest struct {
	OldKey string `json:"oldKey"`
	NewKey string `json:"newKey"`
}

type S3ProxyRenameObjectResult struct {
}

type S3ProxyDeleteObjectRequest struct {
	Key string `json:"key"`
}

type S3ProxyDeleteObjectResult struct {
}

type S3ObjectInfo struct {
	Key          string     `json:"key"`
	Size         int64      `json:"size"`
	LastModified *time.Time `json:"lastModified,omitempty"`
	Etag         string     `json:"etag,omitempty"`

	PresignedGetUrl        string    `json:"presignedGetUrl"`
	PresignedGetUrlExpires time.Time `json:"PresignedGetUrlExpires"`
}
