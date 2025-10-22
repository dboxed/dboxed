package boxes

import (
	"bytes"
	"context"
	"io"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/global"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dustin/go-humanize"
	"github.com/klauspost/compress/gzip"
)

func (s *BoxesServer) restGetBoxRunStatus(c context.Context, i *huma_utils.IdByPath) (*huma_utils.JsonBody[models.BoxRunStatus], error) {
	q := querier2.GetQuerier(c)
	w := global.GetWorkspace(c)

	box, err := dmodel.GetBoxById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}

	err = s.checkBoxToken(c, box.ID)
	if err != nil {
		return nil, err
	}

	boxRunStatus, err := dmodel.GetBoxRunStatus(q, box.ID)
	if err != nil {
		return nil, err
	}

	ret := models.BoxRunStatusFromDB(*boxRunStatus)
	return huma_utils.NewJsonBody(*ret), nil
}

func (s *BoxesServer) restUpdateBoxRunStatus(c context.Context, i *huma_utils.IdByPathAndJsonBody[models.UpdateBoxRunStatus]) (*huma_utils.Empty, error) {
	q := querier2.GetQuerier(c)
	w := global.GetWorkspace(c)

	box, err := dmodel.GetBoxById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}

	err = s.checkBoxToken(c, box.ID)
	if err != nil {
		return nil, err
	}

	boxRunStatus, err := dmodel.GetBoxRunStatus(q, box.ID)
	if err != nil {
		return nil, err
	}

	if i.Body.RunStatus != nil {
		if i.Body.RunStatus.RunStatus != nil {
			err = boxRunStatus.UpdateRunStatus(q, i.Body.RunStatus.RunStatus)
			if err != nil {
				return nil, err
			}
		}
		if i.Body.RunStatus.StartTime != nil {
			err = boxRunStatus.UpdateStartTime(q, i.Body.RunStatus.StartTime)
			if err != nil {
				return nil, err
			}
		}

		if i.Body.RunStatus.StopTime != nil {
			err = boxRunStatus.UpdateStopTime(q, i.Body.RunStatus.StopTime)
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

		err = boxRunStatus.UpdateDockerPs(q, i.Body.DockerPs)
		if err != nil {
			return nil, err
		}
	}

	return &huma_utils.Empty{}, nil
}
