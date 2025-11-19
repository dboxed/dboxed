package s3proxy

import (
	"context"
	"log/slog"
	"path"
	"sync"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dboxed/dboxed/pkg/server/auth_middleware"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/server/resources/huma_metadata"
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
	noTxModifier := huma_utils.MetadataModifier(huma_utils.NoTx, true)
	allowBoxTokenModifier := huma_utils.MetadataModifier(huma_metadata.AllowBoxToken, true)

	huma.Post(workspacesGroup, "/s3-buckets/{id}/proxy/list-objects", s.restListObjects, noTxModifier, allowBoxTokenModifier)
	huma.Post(workspacesGroup, "/s3-buckets/{id}/proxy/presign-put", s.restPresignPut, noTxModifier, allowBoxTokenModifier)
	huma.Post(workspacesGroup, "/s3-buckets/{id}/proxy/rename-object", s.restRenameObject, noTxModifier, allowBoxTokenModifier)
	huma.Post(workspacesGroup, "/s3-buckets/{id}/proxy/delete-object", s.restDeleteObject, noTxModifier, allowBoxTokenModifier)

	return nil
}

func (s *S3ProxyServer) handleBase(ctx context.Context, bucketId string) (*dmodel.S3Bucket, *minio.Client, error) {
	q := querier.GetQuerier(ctx)
	w := auth_middleware.GetWorkspace(ctx)
	token := auth_middleware.GetToken(ctx)

	if token != nil {
		if w.ID != token.Workspace {
			return nil, nil, huma.Error403Forbidden("no access to s3 bucket")
		}
	}

	b, err := dmodel.GetS3BucketById(q, &w.ID, bucketId, true)
	if err != nil {
		return nil, nil, err
	}

	key := bucketLocationCacheKey{
		endpoint: b.Endpoint,
		bucket:   b.Bucket,
	}

	cachedRegion, ok := s.bucketLocationCache.Load(key)
	if !ok {
		c, err := s3utils.BuildS3Client(b, "")
		if err != nil {
			return nil, nil, err
		}
		loc, err := c.GetBucketLocation(ctx, b.Bucket)
		if err != nil {
			return nil, nil, err
		}
		cachedRegion = &loc
		s.bucketLocationCache.Store(key, cachedRegion)
	}
	region := cachedRegion.(*string)

	c, err := s3utils.BuildS3Client(b, *region)
	if err != nil {
		return nil, nil, err
	}
	return b, c, nil
}

func (s *S3ProxyServer) restListObjects(ctx context.Context, i *huma_utils.IdByPathAndJsonBody[models.S3ProxyListObjectsRequest]) (*huma_utils.JsonBody[models.S3ProxyListObjectsResult], error) {
	b, c, err := s.handleBase(ctx, i.Id)
	if err != nil {
		return nil, err
	}

	slog.InfoContext(ctx, "restListObjects", slog.Any("listPrefix", i.Body.Prefix))

	ch := c.ListObjects(ctx, b.Bucket, minio.ListObjectsOptions{
		Prefix: i.Body.Prefix,
	})
	defer func() {
		// drain it
		for range ch {
		}
	}()

	rep := models.S3ProxyListObjectsResult{}
	for o := range ch {
		presignedGetUrl, expires, err := s.presignGet(ctx, b, c, o.Key)
		if err != nil {
			return nil, err
		}
		oi := models.S3ObjectInfo{
			Key:          o.Key,
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

func (s *S3ProxyServer) presignGet(ctx context.Context, b *dmodel.S3Bucket, c *minio.Client, key string) (string, time.Time, error) {
	expiry := time.Hour
	expires := time.Now().Add(expiry).Add(time.Second * 15)
	pr, err := c.PresignedGetObject(ctx, b.Bucket, key, expiry, nil)
	if err != nil {
		return "", time.Time{}, err
	}

	return pr.String(), expires, nil
}

func (s *S3ProxyServer) restPresignPut(ctx context.Context, i *huma_utils.IdByPathAndJsonBody[models.S3ProxyPresignPutRequest]) (*huma_utils.JsonBody[models.S3ProxyPresignPutResult], error) {
	b, c, err := s.handleBase(ctx, i.Id)
	if err != nil {
		return nil, err
	}

	key := path.Clean(i.Body.Key)

	expiry := time.Hour
	expires := time.Now().Add(expiry).Add(time.Second * 15)
	pr, err := c.PresignedPutObject(ctx, b.Bucket, key, expiry)
	if err != nil {
		return nil, err
	}

	return huma_utils.NewJsonBody(models.S3ProxyPresignPutResult{
		PresignedUrl: pr.String(),
		Expires:      expires,
	}), nil
}

func (s *S3ProxyServer) restRenameObject(ctx context.Context, i *huma_utils.IdByPathAndJsonBody[models.S3ProxyRenameObjectRequest]) (*huma_utils.JsonBody[models.S3ProxyRenameObjectResult], error) {
	b, c, err := s.handleBase(ctx, i.Id)
	if err != nil {
		return nil, err
	}

	oldKey := path.Clean(i.Body.OldKey)
	newKey := path.Clean(i.Body.NewKey)

	_, err = c.CopyObject(ctx, minio.CopyDestOptions{
		Bucket: b.Bucket,
		Object: newKey,
	}, minio.CopySrcOptions{
		Bucket: b.Bucket,
		Object: oldKey,
	})
	if err != nil {
		return nil, err
	}
	err = c.RemoveObject(ctx, b.Bucket, oldKey, minio.RemoveObjectOptions{})
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

	key := path.Clean(i.Body.Key)

	err = c.RemoveObject(ctx, vp.Bucket, key, minio.RemoveObjectOptions{})
	if err != nil {
		return nil, err
	}
	rep := models.S3ProxyDeleteObjectResult{}
	return huma_utils.NewJsonBody(rep), nil
}
