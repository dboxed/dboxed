package networks

import (
	"context"
	"log/slog"

	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/models"
)

func (s *NetworksServer) restCreateNetworkNetbird(c context.Context, log *slog.Logger, n *dmodel.Network, body *models.CreateNetworkNetbird) error {
	q := querier2.GetQuerier(c)

	apiUrl := "https://api.netbird.io"
	if body.ApiUrl != nil {
		apiUrl = *body.ApiUrl
	}

	log = log.With(slog.Any("apiUrl", apiUrl))

	n.Netbird = &dmodel.NetworkNetbird{
		ID:             querier2.N(n.ID),
		NetbirdVersion: querier2.N(body.NetbirdVersion),
		ApiUrl:         querier2.N(apiUrl),
		ApiAccessToken: body.ApiAccessToken,
	}
	err := n.Netbird.Create(q)
	if err != nil {
		return err
	}

	return nil
}

func (s *NetworksServer) restUpdateNetworkNetbird(c context.Context, log *slog.Logger, n *dmodel.Network, body *models.UpdateNetworkNetbird) error {
	q := querier2.GetQuerier(c)

	if body.NetbirdVersion != nil {
		err := n.Netbird.UpdateNetbirdVersion(q, *body.NetbirdVersion)
		if err != nil {
			return err
		}
	}
	if body.ApiAccessToken != nil {
		t := body.ApiAccessToken
		if *t == "" {
			t = nil
		}
		err := n.Netbird.UpdateApiAccessToken(q, t)
		if err != nil {
			return err
		}
	}

	return nil
}
