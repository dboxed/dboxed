package boxes

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dboxed/dboxed/pkg/server/auth_middleware"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/util"
	"github.com/dustin/go-humanize"
	"github.com/klauspost/compress/gzip"
)

func (s *BoxesServer) restCreateSandbox(c context.Context, i *huma_utils.IdByPathAndJsonBody[models.CreateBoxSandbox]) (*huma_utils.JsonBody[models.BoxSandbox], error) {
	q := querier2.GetQuerier(c)

	err := util.CheckName(i.Body.Hostname)
	if err != nil {
		return nil, err
	}

	box, err := auth_middleware.CheckResourceAccessAndReturn[dmodel.Box](c, dmodel.TokenTypeBox, i.Id)
	if err != nil {
		return nil, err
	}

	if box.CurrentSandboxId != nil {
		return nil, huma.Error400BadRequest("box already has an active sandbox")
	}

	slog.InfoContext(c, "creating box sandbox", "boxId", box.ID, "machineId", i.Body.MachineId, "hostname", i.Body.Hostname)

	sandbox := &dmodel.BoxSandbox{
		OwnedByWorkspaceOrNull: dmodel.OwnedByWorkspaceOrNull{
			WorkspaceID: querier2.N(box.WorkspaceID),
		},
		BoxID:      querier2.N(box.ID),
		MachineId:  querier2.N(i.Body.MachineId),
		Hostname:   querier2.N(i.Body.Hostname),
		StatusTime: util.Ptr(time.Now()),
	}

	err = sandbox.Create(q)
	if err != nil {
		return nil, err
	}

	err = box.UpdateCurrentSandboxId(q, &sandbox.ID.V)
	if err != nil {
		return nil, err
	}

	err = dmodel.BumpChangeSeq(q, box)
	if err != nil {
		return nil, err
	}

	ret := models.BoxSandboxFromDB(*sandbox)
	return huma_utils.NewJsonBody(*ret), nil
}

func (s *BoxesServer) restListSandboxes(c context.Context, i *huma_utils.IdByPath) (*huma_utils.List[models.BoxSandbox], error) {
	q := querier2.GetQuerier(c)

	err := auth_middleware.CheckResourceAccess(c, dmodel.TokenTypeBox, i.Id)
	if err != nil {
		return nil, err
	}

	l, err := dmodel.ListSandboxesByBox(q, i.Id)
	if err != nil {
		return nil, err
	}

	var ret []models.BoxSandbox
	for _, x := range l {
		ret = append(ret, *models.BoxSandboxFromDB(x))
	}

	return huma_utils.NewList(ret, len(ret)), nil
}

type restGetSandboxInput struct {
	huma_utils.IdByPath
	SandboxId string `path:"sandboxId"`
}

func (s *BoxesServer) restGetSandbox(c context.Context, i *restGetSandboxInput) (*huma_utils.JsonBody[models.BoxSandbox], error) {
	q := querier2.GetQuerier(c)

	err := auth_middleware.CheckResourceAccess(c, dmodel.TokenTypeBox, i.Id)
	if err != nil {
		return nil, err
	}

	sandbox, err := dmodel.GetSandboxById(q, nil, &i.Id, i.SandboxId)
	if err != nil {
		return nil, err
	}

	ret := models.BoxSandboxFromDB(*sandbox)
	return huma_utils.NewJsonBody(*ret), nil
}

type restUpdateSandboxInput struct {
	huma_utils.IdByPath
	SandboxId string `path:"sandboxId"`
	Body      models.UpdateBoxSandboxStatus
}

func (s *BoxesServer) restUpdateSandbox(c context.Context, i *restUpdateSandboxInput) (*huma_utils.Empty, error) {
	q := querier2.GetQuerier(c)

	box, err := auth_middleware.CheckResourceAccessAndReturn[dmodel.Box](c, dmodel.TokenTypeBox, i.Id)
	if err != nil {
		return nil, err
	}

	sandbox, err := dmodel.GetSandboxById(q, nil, &box.ID, i.SandboxId)
	if err != nil {
		return nil, err
	}

	var oldStatusTime, newStatusTime time.Time
	if sandbox.StatusTime != nil {
		oldStatusTime = *sandbox.StatusTime
	}

	if i.Body.SandboxStatus != nil {
		err = sandbox.UpdateStatus(q, i.Body.SandboxStatus.RunStatus, i.Body.SandboxStatus.StartTime, i.Body.SandboxStatus.StopTime, i.Body.SandboxStatus.NetworkIp4)
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

		err = sandbox.UpdateDockerPs(q, i.Body.DockerPs)
		if err != nil {
			return nil, err
		}
	}

	if sandbox.StatusTime != nil {
		newStatusTime = *sandbox.StatusTime
	}

	// cause reconciliation on change or after some time passed, but only if this is the status for the current sandbox
	if box.CurrentSandboxId != nil && *box.CurrentSandboxId == sandbox.ID.V && (oldStatusTime != newStatusTime || time.Now().Sub(newStatusTime) >= time.Second*30) {
		err = dmodel.BumpChangeSeq(q, box)
		if err != nil {
			return nil, err
		}
	}

	return &huma_utils.Empty{}, nil
}

type restReleaseSandboxInput struct {
	huma_utils.IdByPath
	SandboxId string `path:"sandboxId"`
}

func (s *BoxesServer) restReleaseSandbox(c context.Context, i *restReleaseSandboxInput) (*huma_utils.Empty, error) {
	q := querier2.GetQuerier(c)

	box, err := auth_middleware.CheckResourceAccessAndReturn[dmodel.Box](c, dmodel.TokenTypeBox, i.Id)
	if err != nil {
		return nil, err
	}

	if box.CurrentSandboxId == nil {
		return &huma_utils.Empty{}, nil
	}

	if *box.CurrentSandboxId != i.SandboxId {
		return nil, huma.Error400BadRequest(fmt.Sprintf("sandbox %s is not the current sandbox", i.SandboxId))
	}

	err = box.UpdateCurrentSandboxId(q, nil)
	if err != nil {
		return nil, err
	}

	return &huma_utils.Empty{}, nil
}
