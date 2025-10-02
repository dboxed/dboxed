package s3proxy

import (
	"context"
	"log/slog"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/global"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/server/s3utils"
	"github.com/minio/minio-go/v7"
)

type S3ProxyServer struct {
	bucketLocationCache sync.Map
}

type bucketLocationCacheKey struct {
	endpoint string
	bucket   string
}

func New() *S3ProxyServer {
	s := &S3ProxyServer{}
	return s
}

func (s *S3ProxyServer) Init(rootGroup huma.API, workspacesGroup huma.API) error {
	huma.Post(workspacesGroup, "/volume-providers/{id}/s3proxy/list-objects", s.restListObjects)
	huma.Post(workspacesGroup, "/volume-providers/{id}/s3proxy/presign-put", s.restPresignPut)
	huma.Post(workspacesGroup, "/volume-providers/{id}/s3proxy/rename-object", s.restRenameObject)
	huma.Post(workspacesGroup, "/volume-providers/{id}/s3proxy/delete-object", s.restDeleteObject)

	return nil
}

func (s *S3ProxyServer) handleBase(ctx context.Context, vpId int64) (*dmodel.VolumeProvider, *minio.Client, error) {
	q := querier.GetQuerier(ctx)
	w := global.GetWorkspace(ctx)

	vp, err := dmodel.GetVolumeProviderById(q, &w.ID, vpId, true)
	if err != nil {
		return nil, nil, err
	}

	if vp.Rustic == nil || vp.Rustic.StorageS3 == nil {
		return nil, nil, huma.Error500InternalServerError("missing S3 config")
	}

	region := vp.Rustic.StorageS3.Region

	if region == nil {
		key := bucketLocationCacheKey{
			endpoint: vp.Rustic.StorageS3.Endpoint.V,
			bucket:   vp.Rustic.StorageS3.Bucket.V,
		}

		cachedRegion, ok := s.bucketLocationCache.Load(key)
		if !ok {
			c, err := s3utils.BuildS3ClientForRegion(vp, "")
			if err != nil {
				return nil, nil, err
			}
			loc, err := c.GetBucketLocation(ctx, vp.Rustic.StorageS3.Bucket.V)
			if err != nil {
				return nil, nil, err
			}
			cachedRegion = &loc
			s.bucketLocationCache.Store(key, cachedRegion)
		}
		region = cachedRegion.(*string)
	}

	c, err := s3utils.BuildS3ClientForRegion(vp, *region)
	if err != nil {
		return nil, nil, err
	}
	return vp, c, nil
}

func (s *S3ProxyServer) restListObjects(ctx context.Context, i *huma_utils.IdByPathAndJsonBody[models.S3ProxyListObjectsRequest]) (*huma_utils.JsonBody[models.S3ProxyListObjectsResult], error) {
	vp, c, err := s.handleBase(ctx, i.Id)
	if err != nil {
		return nil, err
	}

	slog.InfoContext(ctx, "restListObjects", slog.Any("repoPrefix", vp.Rustic.StorageS3.Prefix.V), slog.Any("listPrefix", i.Body.Prefix))

	prefix := path.Join(vp.Rustic.StorageS3.Prefix.V, i.Body.Prefix)
	if strings.HasSuffix(i.Body.Prefix, "/") {
		prefix += "/"
	}

	ch := c.ListObjects(ctx, vp.Rustic.StorageS3.Bucket.V, minio.ListObjectsOptions{
		Prefix: prefix,
	})
	defer func() {
		// drain it
		for range ch {
		}
	}()

	rep := models.S3ProxyListObjectsResult{}
	for o := range ch {
		presignedGetUrl, expires, err := s.presignGet(ctx, vp, c, o.Key)
		if err != nil {
			return nil, err
		}
		oi := models.S3ObjectInfo{
			Key:          strings.TrimPrefix(o.Key, vp.Rustic.StorageS3.Prefix.V+"/"),
			Size:         o.Size,
			LastModified: &o.LastModified,
			Etag:         o.ETag,

			PresignedGetUrl:        presignedGetUrl,
			PresignedGetUrlExpires: expires,
		}
		rep.Objects = append(rep.Objects, oi)
	}

	return huma_utils.NewJsonBody(rep), nil
}

func (s *S3ProxyServer) presignGet(ctx context.Context, vp *dmodel.VolumeProvider, c *minio.Client, key string) (string, time.Time, error) {
	expiry := time.Hour
	expires := time.Now().Add(expiry).Add(time.Second * 15)
	pr, err := c.PresignedGetObject(ctx, vp.Rustic.StorageS3.Bucket.V, key, expiry, nil)
	if err != nil {
		return "", time.Time{}, err
	}

	return pr.String(), expires, nil
}

func (s *S3ProxyServer) restPresignPut(ctx context.Context, i *huma_utils.IdByPathAndJsonBody[models.S3ProxyPresignPutRequest]) (*huma_utils.JsonBody[models.S3ProxyPresignPutResult], error) {
	vp, c, err := s.handleBase(ctx, i.Id)
	if err != nil {
		return nil, err
	}

	key := path.Join(vp.Rustic.StorageS3.Prefix.V, i.Body.Key)

	expiry := time.Hour
	expires := time.Now().Add(expiry).Add(time.Second * 15)
	pr, err := c.PresignedPutObject(ctx, vp.Rustic.StorageS3.Bucket.V, key, expiry)
	if err != nil {
		return nil, err
	}

	return huma_utils.NewJsonBody(models.S3ProxyPresignPutResult{
		PresignedUrl: pr.String(),
		Expires:      expires,
	}), nil
}

func (s *S3ProxyServer) restRenameObject(ctx context.Context, i *huma_utils.IdByPathAndJsonBody[models.S3ProxyRenameObjectRequest]) (*huma_utils.JsonBody[models.S3ProxyRenameObjectResult], error) {
	vp, c, err := s.handleBase(ctx, i.Id)
	if err != nil {
		return nil, err
	}

	oldKey := path.Join(vp.Rustic.StorageS3.Prefix.V, i.Body.OldKey)
	newKey := path.Join(vp.Rustic.StorageS3.Prefix.V, i.Body.NewKey)

	_, err = c.CopyObject(ctx, minio.CopyDestOptions{
		Bucket: vp.Rustic.StorageS3.Bucket.V,
		Object: newKey,
	}, minio.CopySrcOptions{
		Bucket: vp.Rustic.StorageS3.Bucket.V,
		Object: oldKey,
	})
	if err != nil {
		return nil, err
	}
	err = c.RemoveObject(ctx, vp.Rustic.StorageS3.Bucket.V, oldKey, minio.RemoveObjectOptions{})
	if err != nil {
		return nil, err
	}
	rep := models.S3ProxyRenameObjectResult{}
	return huma_utils.NewJsonBody(rep), nil
}

func (s *S3ProxyServer) restDeleteObject(ctx context.Context, i *huma_utils.IdByPathAndJsonBody[models.S3ProxyDeleteObjectRequest]) (*huma_utils.JsonBody[models.S3ProxyDeleteObjectResult], error) {
	vp, c, err := s.handleBase(ctx, i.Id)
	if err != nil {
		return nil, err
	}

	key := path.Join(vp.Rustic.StorageS3.Prefix.V, i.Body.Key)

	err = c.RemoveObject(ctx, vp.Rustic.StorageS3.Bucket.V, key, minio.RemoveObjectOptions{})
	if err != nil {
		return nil, err
	}
	rep := models.S3ProxyDeleteObjectResult{}
	return huma_utils.NewJsonBody(rep), nil
}
