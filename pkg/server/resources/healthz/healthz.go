package healthz

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dboxed/dboxed/pkg/server/config"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
	"github.com/dboxed/dboxed/pkg/server/models"
	"github.com/dboxed/dboxed/pkg/server/resources/huma_metadata"
)

type HealthzServer struct {
	config config.Config
}

func New(config config.Config) *HealthzServer {
	return &HealthzServer{
		config: config,
	}
}

func (s *HealthzServer) Init(rootGroup huma.API) error {
	huma.Get(rootGroup, "/healthz", s.healthzHandler,
		huma_utils.MetadataModifier(huma_metadata.SkipAuth, true),
		huma_utils.MetadataModifier(huma_utils.NoTx, true),
	)
	return nil
}

func (s *HealthzServer) healthzHandler(ctx context.Context, i *struct{}) (*huma_utils.JsonBody[models.Healthz], error) {
	return huma_utils.NewJsonBody(models.Healthz{Ok: true}), nil
}
