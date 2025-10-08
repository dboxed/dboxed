package server

import (
	"context"
	"net"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
	"github.com/dboxed/dboxed/pkg/server/huma_utils"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func (s *DboxedServer) InitGin() error {
	s.ginEngine = gin.New()

	huma_utils.SetupLogMiddleware(s.ginEngine)
	s.ginEngine.Use(gin.CustomRecovery(func(c *gin.Context, err any) {
		em := huma.ErrorModel{
			Title:  "An internal server error happened",
			Status: 500,
		}
		c.JSON(500, em)
	}))

	corsConf := cors.DefaultConfig()
	corsConf.AllowAllOrigins = true
	corsConf.AllowCredentials = true
	corsConf.AddAllowHeaders("Authorization")
	corsConf.AddExposeHeaders("X-Total-Count")
	s.ginEngine.Use(cors.New(corsConf))

	return nil
}

func (s *DboxedServer) ListenAndServe(ctx context.Context) error {
	server := http.Server{
		Addr:    s.config.Server.ListenAddress,
		Handler: s.ginEngine,
		BaseContext: func(listener net.Listener) context.Context {
			return ctx
		},
	}

	return server.ListenAndServe()
}

func (s *DboxedServer) InitHuma() error {
	s.humaConfig = huma.DefaultConfig("Dboxed API", "0.1.0")

	if s.oidcProvider != nil {
		s.humaConfig.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
			"dboxed": {
				Type: "oauth2",
				Flows: &huma.OAuthFlows{
					Implicit: &huma.OAuthFlow{
						AuthorizationURL: s.oidcProvider.Endpoint().AuthURL,
						TokenURL:         s.oidcProvider.Endpoint().TokenURL,
						Scopes:           map[string]string{},
					},
				},
			},
		}
		s.humaConfig.Security = []map[string][]string{
			{"dboxed": {}},
		}
	}
	s.humaConfig.DocsPath = ""
	s.api = humagin.New(s.ginEngine, s.humaConfig)

	huma_utils.SetupHumaGinContext(s.api)
	huma_utils.SetupTxMiddlewares(s.ginEngine, s.api)
	huma_utils.InitHumaErrorOverride()

	err := huma_utils.InitHumaDocs(s.ginEngine, s.config.Auth.OidcClientId)
	if err != nil {
		return err
	}

	return nil
}
