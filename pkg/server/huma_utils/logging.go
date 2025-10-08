package huma_utils

import (
	"context"
	"fmt"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
	"github.com/gin-gonic/gin"
)

type extraLogValues struct {
	names  []string
	values []any
}

const extraLogValuesContextKey = "extra_log_values"

func ExtraLogValue(ctx any, name string, value any) {
	if ginCtx, ok := ctx.(*gin.Context); ok {
		elvI, ok := ginCtx.Get(extraLogValuesContextKey)
		if !ok {
			elvI = &extraLogValues{}
			ginCtx.Set(extraLogValuesContextKey, elvI)
		}
		elv := elvI.(*extraLogValues)
		elv.names = append(elv.names, name)
		elv.values = append(elv.values, value)
	} else if humaCtx, ok := ctx.(huma.Context); ok {
		ExtraLogValue(humagin.Unwrap(humaCtx), name, value)
	} else if stdCtx, ok := ctx.(context.Context); ok {
		ginCtx := stdCtx.Value("ginContext")
		if ginCtx == nil {
			panic("unknown context type")
		}
		ExtraLogValue(ginCtx, name, value)
	}
}

func SetupLogMiddleware(ginEngine *gin.Engine) {
	loggerConfig := gin.LoggerConfig{
		Output:    gin.DefaultWriter,
		SkipPaths: []string{"/healthz"},
		Formatter: GinLogFormatter,
	}
	ginEngine.Use(gin.LoggerWithConfig(loggerConfig))
}

func SetupHumaGinContext(api huma.API) {
	api.UseMiddleware(func(ctx huma.Context, next func(huma.Context)) {
		ginCtx := humagin.Unwrap(ctx)
		ctx = huma.WithValue(ctx, "ginContext", ginCtx)
		next(ctx)
	})
}

// GinLogFormatter is based on the default logger, but it allows to add custom key/value pairs via the gin context
var GinLogFormatter = func(param gin.LogFormatterParams) string {
	var statusColor, methodColor, resetColor string
	if param.IsOutputColor() {
		statusColor = param.StatusCodeColor()
		methodColor = param.MethodColor()
		resetColor = param.ResetColor()
	}

	if param.Latency > time.Minute {
		param.Latency = param.Latency.Truncate(time.Second)
	}

	var extraKeyValues string

	elvI, ok := param.Keys[extraLogValuesContextKey]
	if ok {
		elv := elvI.(*extraLogValues)
		for i := range elv.names {
			name := elv.names[i]
			value := elv.values[i]

			if len(extraKeyValues) != 0 {
				extraKeyValues += ", "
			}
			extraKeyValues += fmt.Sprintf("%s=%v", name, value)
		}
		extraKeyValues = " | " + extraKeyValues
	}

	return fmt.Sprintf("[GIN] %v |%s %3d %s| %13v | %15s |%s %-7s %s %#v%s\n%s",
		param.TimeStamp.Format("2006/01/02 - 15:04:05"),
		statusColor, param.StatusCode, resetColor,
		param.Latency,
		param.ClientIP,
		methodColor, param.Method, resetColor,
		param.Path,
		extraKeyValues,
		param.ErrorMessage,
	)
}
