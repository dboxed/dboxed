package boxes

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/sse"
	"github.com/dboxed/dboxed/pkg/boxspec"
	"github.com/dboxed/dboxed/pkg/runner/logs/multitail"
	"github.com/dboxed/dboxed/pkg/server/config"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/global"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/server/resources/auth"
	"github.com/dboxed/dboxed/pkg/server/resources/huma_metadata"
	"github.com/dboxed/dboxed/pkg/util"
	"github.com/google/uuid"
	"github.com/nats-io/nkeys"
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
	huma.Get(workspacesGroup, "/boxes/by-uuid/{uuid}", s.restGetBoxByUuid, allowBoxTokenModifier)
	huma.Get(workspacesGroup, "/boxes/by-name/{name}", s.restGetBoxByName, allowBoxTokenModifier)
	huma.Get(workspacesGroup, "/boxes/{id}/box-spec", s.restGetBoxSpec, allowBoxTokenModifier)
	huma.Patch(workspacesGroup, "/boxes/{id}", s.restUpdateBox)
	huma.Delete(workspacesGroup, "/boxes/{id}", s.restDeleteBox)

	// volume attach/detach
	huma.Get(workspacesGroup, "/boxes/{id}/volumes", s.restListAttachedVolumes, allowBoxTokenModifier)
	huma.Post(workspacesGroup, "/boxes/{id}/volumes", s.restAttachVolume)
	huma.Patch(workspacesGroup, "/boxes/{id}/volumes/{volumeId}", s.restUpdateAttachedVolume)
	huma.Delete(workspacesGroup, "/boxes/{id}/volumes/{volumeId}", s.restDetachVolume)

	huma.Get(workspacesGroup, "/boxes/{id}/logs", s.restListLogs)
	sse.Register(workspacesGroup, huma.Operation{
		OperationID: "logs-stream",
		Method:      http.MethodGet,
		Path:        "/boxes/{id}/logs/stream",
		Metadata: map[string]any{
			huma_utils.NoTx: true,
		},
	}, map[string]any{
		"metadata": multitail.LogMetadata{},
		"logs":     boxspec.LogsBatch{},
		"error":    models.LogsError{},
	}, s.sseLogsStream)

	return nil
}

func (s *BoxesServer) restCreateBox(c context.Context, i *huma_utils.JsonBody[models.CreateBox]) (*huma_utils.JsonBody[models.Box], error) {
	q := querier2.GetQuerier(c)

	box, inputErr, err := s.createBox(c, i.Body)
	if err != nil {
		return nil, err
	}
	if inputErr != "" {
		return nil, huma.Error400BadRequest(inputErr)
	}

	ret, err := models.BoxFromDB(c, *box)
	if err != nil {
		return nil, err
	}

	err = dmodel.AddChangeTracking(q, box)
	if err != nil {
		return nil, err
	}

	return huma_utils.NewJsonBody(*ret), nil
}

func (s *BoxesServer) createBox(c context.Context, body models.CreateBox) (*dmodel.Box, string, error) {
	q := querier2.GetQuerier(c)
	w := global.GetWorkspace(c)

	err := util.CheckName(body.Name)
	if err != nil {
		return nil, err.Error(), nil
	}

	nkeyPair, err := nkeys.CreateUser()
	if err != nil {
		return nil, "", err
	}
	nkeySeed, err := nkeyPair.Seed()
	if err != nil {
		return nil, "", err
	}
	nkeyPub, err := nkeyPair.PublicKey()
	if err != nil {
		return nil, "", err
	}

	var networkId *int64
	var networkType *string
	if body.Network != nil {
		var network *dmodel.Network
		network, err = dmodel.GetNetworkById(q, &w.ID, *body.Network, true)
		if err != nil {
			return nil, "", err
		}
		networkId = &network.ID
		networkType = &network.Type
	}

	defaultBoxSpec := boxspec.BoxSpec{
		Dns: &boxspec.DnsSpec{
			Hostname: body.Name,
		},
	}
	b, err := models.MarshalBoxSpec(&defaultBoxSpec)
	if err != nil {
		return nil, "", err
	}

	box := &dmodel.Box{
		OwnedByWorkspace: dmodel.OwnedByWorkspace{
			WorkspaceID: w.ID,
		},
		Uuid: uuid.NewString(),
		Name: body.Name,

		DboxedVersion: "nightly",
		BoxSpec:       b,

		NetworkID:   networkId,
		NetworkType: networkType,

		Nkey:     nkeyPub,
		NkeySeed: string(nkeySeed),
	}

	err = box.Create(q)
	if err != nil {
		return nil, "", err
	}

	if networkId != nil {
		switch global.NetworkType(*networkType) {
		case global.NetworkNetbird:
			box.Netbird = &dmodel.BoxNetbird{
				ID: querier2.N(box.ID),
			}
			err = box.Netbird.Create(q)
			if err != nil {
				return nil, "", err
			}
		default:
			return nil, "unknown network type", nil
		}
	}

	return box, "", nil
}

func (s *BoxesServer) restListBoxes(c context.Context, i *struct{}) (*huma_utils.List[models.Box], error) {
	q := querier2.GetQuerier(c)
	w := global.GetWorkspace(c)
	token := auth.GetToken(c)

	var l []dmodel.Box
	if token != nil && token.BoxID != nil {
		b, err := dmodel.GetBoxById(q, &w.ID, *token.BoxID, true)
		if err != nil {
			return nil, err
		}
		l = append(l, *b)
	} else {
		var err error
		l, err = dmodel.ListBoxesForWorkspace(q, w.ID, true)
		if err != nil {
			return nil, err
		}
	}

	var ret []models.Box
	for _, box := range l {
		mm, err := s.postprocessBox(c, box)
		if err != nil {
			return nil, err
		}
		ret = append(ret, *mm)
	}
	return huma_utils.NewList(ret, len(ret)), nil
}

func (s *BoxesServer) getBoxHelper(c context.Context, box *dmodel.Box) (*huma_utils.JsonBody[models.Box], error) {
	token := auth.GetToken(c)

	if token != nil && token.BoxID != nil {
		if *token.BoxID != box.ID {
			return nil, huma.Error403Forbidden("no access to box")
		}
	}

	boxm, err := s.postprocessBox(c, *box)
	if err != nil {
		return nil, err
	}
	return huma_utils.NewJsonBody(*boxm), nil
}

func (s *BoxesServer) restGetBox(c context.Context, i *huma_utils.IdByPath) (*huma_utils.JsonBody[models.Box], error) {
	q := querier2.GetQuerier(c)
	w := global.GetWorkspace(c)

	box, err := dmodel.GetBoxById(q, &w.ID, i.Id, true)
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

	box, err := dmodel.GetBoxByName(q, w.ID, i.BoxName, true)
	if err != nil {
		return nil, err
	}

	return s.getBoxHelper(c, box)
}

type BoxUuid struct {
	BoxUuid string `path:"uuid"`
}

func (s *BoxesServer) restGetBoxByUuid(c context.Context, i *BoxUuid) (*huma_utils.JsonBody[models.Box], error) {
	q := querier2.GetQuerier(c)
	w := global.GetWorkspace(c)

	box, err := dmodel.GetBoxByUuid(q, w.ID, i.BoxUuid, true)
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

	err = s.doUpdateBox(c, box, i.Body)
	if err != nil {
		return nil, err
	}

	m, err := s.postprocessBox(c, *box)
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
	if body.BoxSpec != nil {
		b, err := models.MarshalBoxSpec(body.BoxSpec)
		if err != nil {
			return err
		}
		err = box.UpdateBoxSpec(q, b)
		if err != nil {
			return err
		}

		// check if the resulting box is valid
		// (this will catch errors in regard to auto-added resources live volumes and networks)
		_, err = s.buildBoxSpec(c, box)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *BoxesServer) restDeleteBox(c context.Context, i *huma_utils.IdByPath) (*huma_utils.Empty, error) {
	q := querier2.GetQuerier(c)
	w := global.GetWorkspace(c)

	err := dmodel.SoftDeleteWithConstraintsByIds[dmodel.Box](q, &w.ID, i.Id)
	if err != nil {
		return nil, err
	}

	return &huma_utils.Empty{}, nil
}

func (s *BoxesServer) postprocessBox(c context.Context, box dmodel.Box) (*models.Box, error) {
	ret, err := models.BoxFromDB(c, box)
	if err != nil {
		return nil, err
	}

	return ret, nil
}
