package s3buckets

import (
	"context"
	"fmt"
	"net/url"
	"regexp"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/global"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/server/resources/huma_metadata"
)

type S3BucketsServer struct {
}

func New() *S3BucketsServer {
	s := &S3BucketsServer{}
	return s
}

func (s *S3BucketsServer) Init(rootGroup huma.API, workspacesGroup huma.API) error {
	allowBoxTokenModifier := huma_utils.MetadataModifier(huma_metadata.AllowBoxToken, true)

	huma.Post(workspacesGroup, "/s3-buckets", s.restCreateS3Bucket)
	huma.Get(workspacesGroup, "/s3-buckets", s.restListS3Buckets, allowBoxTokenModifier)
	huma.Get(workspacesGroup, "/s3-buckets/{id}", s.restGetS3Bucket, allowBoxTokenModifier)
	huma.Get(workspacesGroup, "/s3-buckets/by-bucket-name/{bucket}", s.restGetS3BucketByBucketName, allowBoxTokenModifier)
	huma.Patch(workspacesGroup, "/s3-buckets/{id}", s.restUpdateS3Bucket)
	huma.Delete(workspacesGroup, "/s3-buckets/{id}", s.restDeleteS3Bucket)

	return nil
}

func (s *S3BucketsServer) restCreateS3Bucket(ctx context.Context, i *huma_utils.JsonBody[models.CreateS3Bucket]) (*huma_utils.JsonBody[models.S3Bucket], error) {
	q := querier.GetQuerier(ctx)
	w := global.GetWorkspace(ctx)

	r := dmodel.S3Bucket{
		OwnedByWorkspace: dmodel.OwnedByWorkspace{
			WorkspaceID: w.ID,
		},
		Endpoint:        i.Body.Endpoint,
		Bucket:          i.Body.Bucket,
		AccessKeyId:     i.Body.AccessKeyId,
		SecretAccessKey: i.Body.SecretAccessKey,
	}

	err := r.Create(q)
	if err != nil {
		return nil, err
	}

	return huma_utils.NewJsonBody(models.S3BucketFromDB(r)), nil
}

func (s *S3BucketsServer) restListS3Buckets(ctx context.Context, i *struct{}) (*huma_utils.List[models.S3Bucket], error) {
	q := querier.GetQuerier(ctx)
	w := global.GetWorkspace(ctx)

	l, err := dmodel.ListS3Buckets(q, &w.ID, true)
	if err != nil {
		return nil, err
	}

	var ret []models.S3Bucket
	for _, r := range l {
		mm := models.S3BucketFromDB(r)
		ret = append(ret, mm)
	}
	return huma_utils.NewList(ret, len(ret)), nil
}

func (s *S3BucketsServer) restGetS3Bucket(c context.Context, i *huma_utils.IdByPath) (*huma_utils.JsonBody[models.S3Bucket], error) {
	q := querier.GetQuerier(c)
	w := global.GetWorkspace(c)

	r, err := dmodel.GetS3BucketById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}
	m := models.S3BucketFromDB(*r)
	return huma_utils.NewJsonBody(m), nil
}

type BucketName struct {
	Bucket string `path:"bucket"`
}

func (s *S3BucketsServer) restGetS3BucketByBucketName(c context.Context, i *BucketName) (*huma_utils.JsonBody[models.S3Bucket], error) {
	q := querier.GetQuerier(c)
	w := global.GetWorkspace(c)

	r, err := dmodel.GetS3BucketByBucketName(q, w.ID, i.Bucket, true)
	if err != nil {
		if querier.IsSqlNotFoundError(err) {
			return nil, huma.Error404NotFound(fmt.Sprintf("s3 bucket with name '%s' not found", i.Bucket))
		}
		return nil, err
	}
	m := models.S3BucketFromDB(*r)
	return huma_utils.NewJsonBody(m), nil
}

type restUpdateS3BucketInput struct {
	huma_utils.IdByPath
	huma_utils.JsonBody[models.UpdateS3Bucket]
}

func (s *S3BucketsServer) restUpdateS3Bucket(c context.Context, i *restUpdateS3BucketInput) (*huma_utils.JsonBody[models.S3Bucket], error) {
	q := querier.GetQuerier(c)
	w := global.GetWorkspace(c)

	r, err := dmodel.GetS3BucketById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}

	err = s.doUpdateS3Bucket(c, r, i.Body)
	if err != nil {
		return nil, err
	}

	m := models.S3BucketFromDB(*r)

	return huma_utils.NewJsonBody(m), nil
}

func (s *S3BucketsServer) doUpdateS3Bucket(c context.Context, r *dmodel.S3Bucket, body models.UpdateS3Bucket) error {
	q := querier.GetQuerier(c)

	if body.Endpoint != nil {
		err := s.checkEndpoint(*body.Endpoint)
		if err != nil {
			return err
		}
		err = r.UpdateEndpoint(q, *body.Endpoint)
		if err != nil {
			return err
		}
	}
	if body.Bucket != nil {
		err := r.UpdateBucket(q, *body.Bucket)
		if err != nil {
			return err
		}
	}

	if body.AccessKeyId != nil || body.SecretAccessKey != nil {
		if body.AccessKeyId == nil || body.SecretAccessKey == nil {
			return huma.Error400BadRequest("either all or none of accessKeyId and secretAccessKey must be set")
		}
		err := r.UpdateKeys(q, *body.AccessKeyId, *body.SecretAccessKey)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *S3BucketsServer) restDeleteS3Bucket(c context.Context, i *huma_utils.IdByPath) (*huma_utils.Empty, error) {
	q := querier.GetQuerier(c)
	w := global.GetWorkspace(c)

	err := dmodel.SoftDeleteWithConstraintsByIds[*dmodel.S3Bucket](q, &w.ID, i.Id)
	if err != nil {
		return nil, err
	}

	return &huma_utils.Empty{}, nil
}

func (s *S3BucketsServer) checkEndpoint(endpoint string) error {
	u, err := url.Parse(endpoint)
	if err != nil {
		return huma.Error400BadRequest("invalid endpoint", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return huma.Error400BadRequest("invalid endpoint scheme")
	}
	return nil
}

var prefixRegex = regexp.MustCompile(`^([-a-zA-Z0-9]*)(/([-a-zA-Z0-9]+))*/?$`)

func (s *S3BucketsServer) checkPrefix(prefix string) error {
	if !prefixRegex.MatchString(prefix) {
		return fmt.Errorf("invalid prefix")
	}
	return nil
}
