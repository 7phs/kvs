package server

import (
	"github.com/7phs/kvs/internal/config"
	"github.com/7phs/kvs/internal/storages"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"net/http"
	"sync"
)

type Server interface {
	Start() error
	Stop()
}

type DefaultServer struct {
	logger      *zap.Logger
	storages    storages.Storages
	maintenance storages.Maintenance
	server      http.Server
}

func NewServer(
	logger *zap.Logger,
	config config.Config,
	storages storages.Storages,
	maintenance storages.Maintenance,
) Server {
	return &DefaultServer{}
}

func (o *DefaultServer) Start() error {
	var wg sync.WaitGroup

	go o.maintenance.Start(ctx)

	go o.stop(ctx)
}

func (o *DefaultServer) Stop() {
	var wg errgroup.Group

	wg.Go(o.server.Shutdown(ctx))

	wg.Go(o.maintenance.Shutdown())

	return wg.Wait()
}
