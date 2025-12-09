package networks

import (
	"context"
	"log/slog"
	"net/url"
	"regexp"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dboxed/dboxed/pkg/server/db/dmodel"
	querier2 "github.com/dboxed/dboxed/pkg/server/db/querier"
	"github.com/dboxed/dboxed/pkg/server/models"
)

var netbirdVersionRegex = regexp.MustCompile(`^(latest|[0-9]+\.[0-9]+(\.[0-9]+)?)$`)

func (s *NetworksServer) restCreateNetworkNetbird(c context.Context, log *slog.Logger, n *dmodel.Network, body *models.CreateNetworkNetbird) error {
	q := querier2.GetQuerier(c)

	apiUrl := "https://api.netbird.io"
	if body.ApiUrl != nil {
		apiUrl = *body.ApiUrl
	}
	if apiUrl == "" {
		return huma.Error400BadRequest("api url can't be empty")
	}
	_, err := url.Parse(apiUrl)
	if err != nil {
		return huma.Error400BadRequest("invalid api url", err)
	}

	if body.ApiAccessToken == "" {
		return huma.Error400BadRequest("api access token can't be empty")
	}
	if !netbirdVersionRegex.MatchString(body.NetbirdVersion) {
		return huma.Error400BadRequest("invalid netbird version")
	}

	log = log.With(slog.Any("apiUrl", apiUrl))

	n.Netbird = &dmodel.NetworkNetbird{
		ID:             querier2.N(n.ID),
		NetbirdVersion: querier2.N(body.NetbirdVersion),
		ApiUrl:         querier2.N(apiUrl),
		ApiAccessToken: querier2.N(body.ApiAccessToken),
	}
	err = n.Netbird.Create(q)
	if err != nil {
		return err
	}

	return nil
}

func (s *NetworksServer) restUpdateNetworkNetbird(c context.Context, log *slog.Logger, n *dmodel.Network, body *models.UpdateNetworkNetbird) error {
	q := querier2.GetQuerier(c)

	if body.NetbirdVersion != nil {
		if !netbirdVersionRegex.MatchString(*body.NetbirdVersion) {
			return huma.Error400BadRequest("invalid netbird version")
		}
		err := n.Netbird.UpdateNetbirdVersion(q, *body.NetbirdVersion)
		if err != nil {
			return err
		}
	}
	if body.ApiAccessToken != nil {
		err := n.Netbird.UpdateApiAccessToken(q, *body.ApiAccessToken)
		if err != nil {
			return err
		}
	}

	return nil
}
