package huma_utils

import (
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

func InitHumaDocs(ginEngine *gin.Engine, oidcClientId string) error {
	unpkUrl, err := url.Parse("https://unpkg.com/swagger-ui-dist@5.22.0")
	if err != nil {
		return err
	}
	proxy := &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.Out.URL.Path = strings.TrimPrefix(r.Out.URL.Path, "/docs/")
			r.SetURL(unpkUrl)
			r.Out.Host = unpkUrl.Host
		},
	}
	ginEngine.GET("/docs/*any", func(c *gin.Context) {
		if c.Request.URL.Path == "/docs" || c.Request.URL.Path == "/docs/" || c.Request.URL.Path == "/docs/index.html" {
			c.Header("Content-Type", "text/html")
			_, _ = c.Writer.WriteString(buildSwaggerIndexHtml(oidcClientId))
		} else {
			proxy.ServeHTTP(c.Writer, c.Request)
		}
	})

	return nil
}

func buildSwaggerIndexHtml(oauth2ClientId string) string {
	return `<!DOCTYPE html>
	<html lang="en">
	<head>
	  <meta charset="utf-8" />
	  <meta name="viewport" content="width=device-width, initial-scale=1" />
	  <meta name="description" content="SwaggerUI" />
	  <title>SwaggerUI</title>
	  <link rel="stylesheet" href="./swagger-ui.css" />
	</head>
	<body>
	<div id="swagger-ui"></div>
	<script src="./swagger-ui-bundle.js" crossorigin></script>
	<script>
	  window.onload = () => {
	    window.ui = SwaggerUIBundle({
	      url: '/openapi.json',
	      dom_id: '#swagger-ui',
	      oauth2: {
	        client_id: "` + oauth2ClientId + `"
	      }
	    });
	  };
	</script>
	</body>
	</html>
`
}
