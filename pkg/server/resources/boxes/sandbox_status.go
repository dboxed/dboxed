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

	box, err := dmodel.GetBoxWithSandboxStatusById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}

	err = s.checkBoxToken(c, box.ID)
	if err != nil {
		return nil, err
	}

	ret := models.BoxSandboxStatusFromDB(*box.SandboxStatus)
	return huma_utils.NewJsonBody(*ret), nil
}

func (s *BoxesServer) restUpdateSandboxStatus(c context.Context, i *huma_utils.IdByPathAndJsonBody[models.UpdateBoxSandboxStatus]) (*huma_utils.Empty, error) {
	q := querier2.GetQuerier(c)
	w := auth_middleware.GetWorkspace(c)

	box, err := dmodel.GetBoxWithSandboxStatusById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}

	err = s.checkBoxToken(c, box.ID)
	if err != nil {
		return nil, err
	}

	oldStatusTime := box.SandboxStatus.StatusTime

	if i.Body.SandboxStatus != nil {
		if i.Body.SandboxStatus.RunStatus != nil {
			err = box.SandboxStatus.UpdateRunStatus(q, i.Body.SandboxStatus.RunStatus)
			if err != nil {
				return nil, err
			}
		}
		if i.Body.SandboxStatus.StartTime != nil {
			err = box.SandboxStatus.UpdateStartTime(q, i.Body.SandboxStatus.StartTime)
			if err != nil {
				return nil, err
			}
		}

		if i.Body.SandboxStatus.StopTime != nil {
			err = box.SandboxStatus.UpdateStopTime(q, i.Body.SandboxStatus.StopTime)
			if err != nil {
				return nil, err
			}
		}

		if i.Body.SandboxStatus.NetworkIp4 != nil {
			err = box.SandboxStatus.UpdateNetworkIp4(q, i.Body.SandboxStatus.NetworkIp4)
			if err != nil {
				return nil, err
			}
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

	if oldStatusTime != nil && box.SandboxStatus.StatusTime != nil {
		// if we didn't update status for some time, do immediate reconciliation so that the overall box status gets
		// updates asap
		if box.SandboxStatus.StatusTime.Sub(*oldStatusTime) >= time.Second*30 {
			err = dmodel.AddChangeTracking(q, box)
			if err != nil {
				return nil, err
			}
		}
	}

	return &huma_utils.Empty{}, nil
}
