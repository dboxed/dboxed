package boxes

import (
	"bytes"
	"context"
	"io"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dboxed/dboxed/pkg/server/auth_middleware"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dustin/go-humanize"
	"github.com/klauspost/compress/gzip"
)

func (s *BoxesServer) restGetSandboxStatus(c context.Context, i *huma_utils.IdByPath) (*huma_utils.JsonBody[models.BoxSandboxStatus], error) {
	q := querier2.GetQuerier(c)
	w := auth_middleware.GetWorkspace(c)

	err := auth_middleware.CheckTokenAccess(c, dmodel.TokenTypeBox, i.Id)
	if err != nil {
		return nil, err
	}

	box, err := dmodel.GetBoxWithSandboxStatusById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}

	ret := models.BoxSandboxStatusFromDB(*box.SandboxStatus)
	return huma_utils.NewJsonBody(*ret), nil
}

func (s *BoxesServer) restUpdateSandboxStatus(c context.Context, i *huma_utils.IdByPathAndJsonBody[models.UpdateBoxSandboxStatus]) (*huma_utils.Empty, error) {
	q := querier2.GetQuerier(c)
	w := auth_middleware.GetWorkspace(c)

	err := auth_middleware.CheckTokenAccess(c, dmodel.TokenTypeBox, i.Id)
	if err != nil {
		return nil, err
	}

	box, err := dmodel.GetBoxWithSandboxStatusById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}

	var oldStatusTime, newStatusTime time.Time
	if box.SandboxStatus.StatusTime != nil {
		oldStatusTime = *box.SandboxStatus.StatusTime
	}

	if i.Body.SandboxStatus != nil {
		err = box.SandboxStatus.UpdateStatus(q, i.Body.SandboxStatus.RunStatus, i.Body.SandboxStatus.StartTime, i.Body.SandboxStatus.StopTime, i.Body.SandboxStatus.NetworkIp4)
		if err != nil {
			return nil, err
		}
	}

	if i.Body.DockerPs != nil {
		if len(i.Body.DockerPs) > humanize.KiByte*16 {
			return nil, huma.Error400BadRequest("dockerPs is too large")
		}
		r, err := gzip.NewReader(bytes.NewReader(i.Body.DockerPs))
		if err != nil {
			return nil, huma.Error400BadRequest("invalid dockerPs", err)
		}
		defer r.Close()
		w, err := io.Copy(io.Discard, r)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid dockerPs", err)
		}
		if w > humanize.KiByte*64 {
			return nil, huma.Error400BadRequest("extracted dockerPs is too large")
		}

		err = box.SandboxStatus.UpdateDockerPs(q, i.Body.DockerPs)
		if err != nil {
			return nil, err
		}
	}

	if box.SandboxStatus.StatusTime != nil {
		newStatusTime = *box.SandboxStatus.StatusTime
	}

	// cause reconciliation on change or after some time passed
	if oldStatusTime != newStatusTime || time.Now().Sub(newStatusTime) >= time.Second*30 {
		err = dmodel.BumpChangeSeq(q, box)
		if err != nil {
			return nil, err
		}
	}

	return &huma_utils.Empty{}, nil
}
