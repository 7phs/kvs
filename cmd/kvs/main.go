package main

import (
	"context"
	"github.com/7phs/kvs/internal/config"
	"github.com/7phs/kvs/internal/server"
	"github.com/7phs/kvs/internal/storages"
	"go.uber.org/zap"
	"log"
	"os"
	"os/signal"
)

func main() {
	conf, err := config.NewConfigFromEnv()
	if err != nil {
		log.Fatal("failed to prepare config: %w", err)
	}

	logger, err := zap.NewProduction() // TODO: setup log level
	if err != nil {
		log.Fatal("failed to init logger: %w", err)
	}
	defer logger.Sync()

	logger.Info("APP RUN")

	logger.Info("config",
		zap.Int(config.PORT, conf.Port()),
		zap.Duration(config.EXPIRATION, conf.Expiration()),
		zap.Duration(config.MAINTENANCE, conf.Maintenance()),
	)

	logger.Info("init: data pool")

	dataPool, err := storages.NewDataPool()
	if err != nil {
		logger.Fatal("failed to init data pool",
			zap.Error(err),
		)
	}

	logger.Info("init: data dictionary")

	dataDictionary := storages.NewDataChunks()

	logger.Info("init: maintainer")

	maintainer := storages.NewGroupMaintenance(logger,
		&dataPool,
		&dataDictionary,
	)

	logger.Info("init: storages")

	storages, err := storages.NewInMemStorages(
		conf,
		&dataPool,
		&dataDictionary,
	)
	if err != nil {
		logger.Fatal("failed to init data pool",
			zap.Error(err),
		)
	}

	logger.Info("init: server")

	srv := server.NewServer(
		logger,
		conf,
		storages,
		maintainer,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		logger.Info("interrupt")

		cancel()
	}()

	go func() {
		logger.Info("start: server")

		err := srv.Start()
		if err != nil {
			logger.Fatal("failed to start server",
				zap.Error(err),
			)
			os.Exit(2)
		}
	}()

	<-ctx.Done()

	logger.Info("stop: server")

	srv.Stop()

	logger.Info("finish")
}
