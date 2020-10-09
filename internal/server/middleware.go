package server

import (
	"time"

	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

func NewLoggerHandler(logger *zap.Logger, handler fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		method := string(ctx.Method())
		begin := time.Now()

		defer logger.Debug(method,
			zap.String("url", string(ctx.RequestURI())),
			zap.Int("status", ctx.Response.StatusCode()),
			zap.Duration("elapse", time.Since(begin)),
		)

		handler(ctx)
	}
}
