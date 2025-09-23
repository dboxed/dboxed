package boxes

import (
	"context"
	"time"

	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	"github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/global"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
	"github.com/dboxed/dboxed/pkg/server/models"
)

func (s *BoxesServer) restGenerateToken(c context.Context, i *huma_utils.IdByPath) (*huma_utils.JsonBody[models.BoxToken], error) {
	q := querier.GetQuerier(c)
	w := global.GetWorkspace(c)

	box, err := dmodel.GetBoxById(q, &w.ID, i.Id, true)
	if err != nil {
		return nil, err
	}

	token := models.OldToken{
		TokenVersion: models.CurrentTokenVersion,
		WorkspaceId:  w.ID,
		BoxId:        box.ID,
		Exp:          time.Now().Add(time.Hour).Unix(),
	}
	tokenStr, err := token.BuildTokenStr([]byte(box.NkeySeed))
	if err != nil {
		return nil, err
	}

	return huma_utils.NewJsonBody(models.BoxToken{
		Token: tokenStr,
	}), nil
}
