package boxes

import (
	"context"
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/sse"
	"github.com/dboxed/dboxed/pkg/boxspec"
	"github.com/dboxed/dboxed/pkg/server/config"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/global"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/server/resources/auth"
	"github.com/dboxed/dboxed/pkg/server/resources/boxes_utils"
	"github.com/dboxed/dboxed/pkg/server/resources/huma_metadata"
)

type BoxesServer struct {
	config config.Config
}

func New(config config.Config) *BoxesServer {
	return &BoxesServer{
		config: config,
	}
}

func (s *BoxesServer) Init(rootGroup huma.API, workspacesGroup huma.API) error {
	allowBoxTokenModifier := huma_utils.MetadataModifier(huma_metadata.AllowBoxToken, true)

	huma.Post(workspacesGroup, "/boxes", s.restCreateBox)
	huma.Get(workspacesGroup, "/boxes", s.restListBoxes, allowBoxTokenModifier)
	huma.Get(workspacesGroup, "/boxes/{id}", s.restGetBox, allowBoxTokenModifier)
	huma.Get(workspacesGroup, "/boxes/by-name/{name}", s.restGetBoxByName, allowBoxTokenModifier)
	huma.Get(workspacesGroup, "/boxes/{id}/box-spec", s.restGetBoxSpec, allowBoxTokenModifier)
	huma.Patch(workspacesGroup, "/boxes/{id}", s.restUpdateBox)
	huma.Delete(workspacesGroup, "/boxes/{id}", s.restDeleteBox)

	// compose-projects
	huma.Get(workspacesGroup, "/boxes/{id}/compose-projects", s.restListComposeProjects, allowBoxTokenModifier)
	huma.Post(workspacesGroup, "/boxes/{id}/compose-projects", s.restCreateComposeProject)
	huma.Patch(workspacesGroup, "/boxes/{id}/compose-projects/{composeName}", s.restUpdateComposeProject)
	huma.Delete(workspacesGroup, "/boxes/{id}/compose-projects/{composeName}", s.restDeleteComposeProject)

	// volume attach/detach
	huma.Get(workspacesGroup, "/boxes/{id}/volumes", s.restListAttachedVolumes, allowBoxTokenModifier)
	huma.Post(workspacesGroup, "/boxes/{id}/volumes", s.restAttachVolume)
	huma.Patch(workspacesGroup, "/boxes/{id}/volumes/{volumeId}", s.restUpdateAttachedVolume)
	huma.Delete(workspacesGroup, "/boxes/{id}/volumes/{volumeId}", s.restDetachVolume)

	// port forwards
	huma.Get(workspacesGroup, "/boxes/{id}/port-forwards", s.restListPortForwards, allowBoxTokenModifier)
	huma.Post(workspacesGroup, "/boxes/{id}/port-forwards", s.restCreatePortForward)
	huma.Patch(workspacesGroup, "/boxes/{id}/port-forwards/{portForwardId}", s.restUpdatePortForward)
	huma.Delete(workspacesGroup, "/boxes/{id}/port-forwards/{portForwardId}", s.restDeletePortForward)

	// ingresses
	huma.Get(workspacesGroup, "/boxes/{id}/ingresses", s.restListBoxIngresses, allowBoxTokenModifier)
	huma.Post(workspacesGroup, "/boxes/{id}/ingresses", s.restCreateBoxIngress)
	huma.Patch(workspacesGroup, "/boxes/{id}/ingresses/{ingressId}", s.restUpdateBoxIngress)
	huma.Delete(workspacesGroup, "/boxes/{id}/ingresses/{ingressId}", s.restDeleteBoxIngress)

	// run status
	huma.Get(workspacesGroup, "/boxes/{id}/sandbox-status", s.restGetSandboxStatus, allowBoxTokenModifier)
	huma.Patch(workspacesGroup, "/boxes/{id}/sandbox-status", s.restUpdateSandboxStatus, allowBoxTokenModifier)

	// logs
	huma.Post(workspacesGroup, "/boxes/{id}/logs", s.restPostLogs, allowBoxTokenModifier)
	huma.Get(workspacesGroup, "/boxes/{id}/logs", s.restListLogs, allowBoxTokenModifier)
	sse.Register(workspacesGroup, huma.Operation{
		OperationID: "logs-stream",
		Method:      http.MethodGet,
		Path:        "/boxes/{id}/logs/{logId}/stream",
		Metadata: map[string]any{
			huma_utils.NoTx: true,
		},
	}, map[string]any{
		"metadata":       models.LogMetadataModel{},
		"logs-batch":     boxspec.LogsBatch{},
		"end-of-history": endOfHistory{},
		"error":          models.LogsError{},
	}, s.sseLogsStream)

	return nil
}

func (s *BoxesServer) restCreateBox(c context.Context, i *huma_utils.JsonBody[models.CreateBox]) (*huma_utils.JsonBody[models.Box], error) {
	box, inputErr, err := boxes_utils.CreateBox(c, i.Body, global.BoxTypeNormal)
	if err != nil {
		return nil, err
	}
	if inputErr != "" {
		return nil, huma.Error400BadRequest(inputErr)
	}

	ret, err := models.BoxFromDB(*box, nil)
	if err != nil {
		return nil, err
	}

	return huma_utils.NewJsonBody(*ret), nil
}

func (s *BoxesServer) restListBoxes(c context.Context, i *struct{}) (*huma_utils.List[models.Box], error) {
	q := querier2.GetQuerier(c)
	w := global.GetWorkspace(c)
	token := auth.GetToken(c)

	var l []dmodel.BoxWithSandboxStatus
	if token != nil && token.BoxID != nil {
		b, err := dmodel.GetBoxWithSandboxStatusById(q, &w.ID, *token.BoxID, true)
		if err != nil {
			return nil, err
		}
		l = append(l, *b)
	} else {
		var err error
		l, err = dmodel.ListBoxesWithSandboxStatusForWorkspace(q, w.ID, true)
		if err != nil {
			return nil, err
		}
	}

	var ret []models.Box
	for _, box := range l {
		mm, err := s.postprocessBox(box.Box, box.SandboxStatus)
		if err != nil {
			return nil, err
		}
		ret = append(ret, *mm)
	}
	return huma_utils.NewList(ret, len(ret)), nil
}

func (s *BoxesServer) checkBoxToken(c context.Context, boxId string) error {
	token := auth.GetToken(c)

	if token != nil && token.BoxID != nil {
		if *token.BoxID != boxId {
			return huma.Error403Forbidden("no access to box")
		}
	}

	return nil
}

func (s *BoxesServer) getBoxHelper(c context.Context, box *dmodel.BoxWithSandboxStatus) (*huma_utils.JsonBody[models.Box], error) {
	err := s.checkBoxToken(c, box.ID)
	if err != nil {
		return nil, err
	}

	boxm, err := s.postprocessBox(box.Box, box.SandboxStatus)
	if err != nil {
		return nil, err
	}
	return huma_utils.NewJsonBody(*boxm), nil
}

func (s *BoxesServer) restGetBox(c context.Context, i *huma_utils.IdByPath) (*huma_utils.JsonBody[models.Box], error) {
	q := querier2.GetQuerier(c)
	w := global.GetWorkspace(c)

	box, err := dmodel.GetBoxWithSandboxStatusById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}

	return s.getBoxHelper(c, box)
}

type BoxName struct {
	BoxName string `path:"name"`
}

func (s *BoxesServer) restGetBoxByName(c context.Context, i *BoxName) (*huma_utils.JsonBody[models.Box], error) {
	q := querier2.GetQuerier(c)
	w := global.GetWorkspace(c)

	box, err := dmodel.GetBoxWithSandboxStatusByName(q, w.ID, i.BoxName, true)
	if err != nil {
		return nil, err
	}

	return s.getBoxHelper(c, box)
}

type restUpdateBoxInput struct {
	huma_utils.IdByPath
	huma_utils.JsonBody[models.UpdateBox]
}

func (s *BoxesServer) restUpdateBox(c context.Context, i *restUpdateBoxInput) (*huma_utils.JsonBody[models.Box], error) {
	q := querier2.GetQuerier(c)
	w := global.GetWorkspace(c)

	box, err := dmodel.GetBoxById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}
	if err = s.checkNormalBoxMod(box); err != nil {
		return nil, err
	}

	err = s.doUpdateBox(c, box, i.Body)
	if err != nil {
		return nil, err
	}

	m, err := s.postprocessBox(*box, nil)
	if err != nil {
		return nil, err
	}

	err = dmodel.AddChangeTracking(q, box)
	if err != nil {
		return nil, err
	}

	return huma_utils.NewJsonBody(*m), nil
}

func (s *BoxesServer) doUpdateBox(c context.Context, box *dmodel.Box, body models.UpdateBox) error {
	q := querier2.GetQuerier(c)

	if body.DesiredState != nil {
		// Validate desired state
		desiredState := *body.DesiredState
		if desiredState != "up" && desiredState != "down" {
			return huma.Error400BadRequest("desiredState must be either 'up' or 'down'")
		}

		// Update the desired state
		err := box.UpdateDesiredState(q, desiredState)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *BoxesServer) restDeleteBox(c context.Context, i *huma_utils.IdByPath) (*huma_utils.Empty, error) {
	q := querier2.GetQuerier(c)
	w := global.GetWorkspace(c)

	box, err := dmodel.GetBoxById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}
	if err = s.checkNormalBoxMod(box); err != nil {
		return nil, err
	}

	v, err := dmodel.ListVolumesByMountBoxId(q, &w.ID, box.ID, false) // we must also include deleted volumes
	if err != nil {
		return nil, err
	}
	for _, x := range v {
		if x.MountStatus.MountId.Valid && x.MountStatus.ReleaseTime == nil {
			return nil, huma.Error400BadRequest("can't delete box while a volume is mounted by the box")
		}
	}

	err = dmodel.SoftDeleteWithConstraintsByIds[*dmodel.Box](q, &w.ID, i.Id)
	if err != nil {
		return nil, err
	}

	return &huma_utils.Empty{}, nil
}

func (s *BoxesServer) postprocessBox(box dmodel.Box, sandboxStatus *dmodel.BoxSandboxStatus) (*models.Box, error) {
	ret, err := models.BoxFromDB(box, sandboxStatus)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (s *BoxesServer) checkNormalBoxMod(box *dmodel.Box) error {
	if box.BoxType != string(global.BoxTypeNormal) {
		return huma.Error400BadRequest(fmt.Sprintf("modifications on %s boxes not allowed", box.BoxType))
	}
	return nil
}
