package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/7phs/kvs/internal/config"
	"github.com/7phs/kvs/internal/storages"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type Server interface {
	Start() error
	Stop()
}

type DefaultServer struct {
	logger              *zap.Logger
	maintenance         GroupMaintenance
	port                int
	maintenanceInterval time.Duration
	server              fasthttp.Server

	cancelCtx context.Context
	cancel    func()

	storages storages.Storages
}

func NewServer(
	logger *zap.Logger,
	conf config.Config,
	storages storages.Storages,
) Server {
	cancelCtx, cancel := context.WithCancel(context.Background())

	srv := &DefaultServer{
		logger:              logger,
		storages:            storages,
		port:                conf.Port(),
		maintenanceInterval: conf.Maintenance(),

		cancelCtx: cancelCtx,
		cancel:    cancel,

		maintenance: NewGroupMaintenance(logger, storages),
	}
	srv.server.Handler = srv.handler

	return srv
}

func (o *DefaultServer) handler(ctx *fasthttp.RequestCtx) {
	switch string(ctx.Method()) {
	case http.MethodGet:
		key := ctx.Path()

		o.logger.Debug("handle GET",
			zap.ByteString("key", key),
		)

		body, err := o.storages.Get(key)
		if err != nil {
			o.handlerError(ctx, err)
			return
		}

		defer body.Free()

		ctx.SetStatusCode(fasthttp.StatusOK)
		ctx.SetBody(body.Bytes())

	case http.MethodPost:
		key := ctx.Path()

		o.logger.Debug("handle POST",
			zap.ByteString("key", key),
		)

		err := o.storages.Add(ctx.Path(), ctx.Request.Body())
		if err != nil {
			o.handlerError(ctx, err)
			return
		}

		ctx.SetStatusCode(fasthttp.StatusOK)

	default:
		ctx.Error("Unsupported method", fasthttp.StatusMethodNotAllowed)
	}
}

func (o *DefaultServer) handlerError(ctx *fasthttp.RequestCtx, err error) {
	switch err {
	case storages.ErrKeyNotFound:
		ctx.Error("Not found", fasthttp.StatusNotFound)
	default:
		ctx.Error("Internal error", fasthttp.StatusInternalServerError)
	}
}

func (o *DefaultServer) Start() error {
	var wg errgroup.Group

	wg.Go(func() error {
		o.logger.Info("maintenance: start")

		o.maintenance.Start(o.cancelCtx, o.maintenanceInterval)
		return nil
	})

	wg.Go(func() error {
		port := fmt.Sprintf(":%d", o.port)

		o.logger.Info("http: listen",
			zap.String("port", port),
		)

		return o.server.ListenAndServe(port)
	})

	return wg.Wait()
}

func (o *DefaultServer) Stop() {
	var wg errgroup.Group

	wg.Go(func() error {
		o.logger.Info("http: shutdown")

		return o.server.Shutdown()
	})

	wg.Go(func() error {
		o.logger.Info("maintenance: shutdown")

		o.cancel()

		return nil
	})

	err := wg.Wait()
	if err != nil {
		o.logger.Error("failed to stop server",
			zap.Error(err),
		)
	}
}
