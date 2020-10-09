package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/7phs/kvs/internal/config"
	"github.com/7phs/kvs/internal/server"
	"github.com/7phs/kvs/internal/storages"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	errExitCode = 2
)

func logLevel(conf config.Config) zap.AtomicLevel {
	var level zapcore.Level

	switch conf.LogLevel() {
	case config.LogLevelDebug:
		level = zapcore.DebugLevel
	case config.LogLevelInfo:
		level = zapcore.InfoLevel
	case config.LogLevelWarning:
		level = zapcore.WarnLevel
	case config.LogLevelError:
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}

	return zap.NewAtomicLevelAt(level)
}

func buildLogger(conf config.Config) (*zap.Logger, error) {
	logConfig := zap.NewProductionConfig()
	logConfig.Level = logLevel(conf)

	return logConfig.Build()
}

func newMapDictionary(conf config.Config) (storages.DataDictionary, error) {
	memoryPool := storages.NewMemoryPool(conf.PreAllocated())

	pool, err := storages.NewDataPool(memoryPool)
	if err != nil {
		return nil, err
	}

	return storages.NewMapDictionary(pool), nil
}

func newSyncMapDictionary(conf config.Config) (storages.DataDictionary, error) {
	memoryPool := storages.NewMemoryPool(conf.PreAllocated())

	pool, err := storages.NewDataPool(memoryPool)
	if err != nil {
		return nil, err
	}

	return storages.NewSyncMapDictionary(pool), nil
}

func initStorages(conf config.Config) (dictionary storages.DataDictionary, err error) {
	switch conf.Mode() {
	case config.StorageModeMap:
		return newMapDictionary(conf)

	case config.StorageModeSyncMap:
		return newSyncMapDictionary(conf)

	case config.StorageModePartitionedMap:
		return storages.NewPartitionedDictionary(
			storages.DefaultPartitionNum,
			storages.DefaultPartitionMask,
			func() (storages.DataDictionary, error) {
				return newMapDictionary(conf)
			},
		)
	case config.StorageModePartitionedSyncMap:
		return storages.NewPartitionedDictionary(
			storages.DefaultPartitionNum,
			storages.DefaultPartitionMask,
			func() (storages.DataDictionary, error) {
				return newSyncMapDictionary(conf)
			},
		)
	}

	return
}

func main() {
	conf, err := config.NewConfigFromEnv()
	if err != nil {
		log.Fatal("failed to prepare config: %w", err)
	}

	logger, err := buildLogger(conf)
	if err != nil {
		log.Fatal("failed to init logger: %w", err)
	}

	defer func() {
		_ = logger.Sync()
	}()

	logger.Info("APP RUN")

	logger.Info("config",
		zap.String(config.LOGLEVEL, string(conf.LogLevel())),
		zap.Int(config.PORT, conf.Port()),
		zap.Duration(config.EXPIRATION, conf.Expiration()),
		zap.Duration(config.MAINTENANCE, conf.Maintenance()),
		zap.Int(config.PREALLOCATED, conf.PreAllocated()),
		zap.String(config.MODE, string(conf.Mode())),
	)

	logger.Info("init: data dictionary")

	dictionary, err := initStorages(conf)
	if err != nil {
		logger.Fatal("failed to init data dictionary")
	}

	logger.Info("init: storages")

	storages, err := storages.NewInMemStorages(
		conf,
		dictionary,
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
			os.Exit(errExitCode)
		}
	}()

	<-ctx.Done()

	logger.Info("stop: server")

	srv.Stop()

	logger.Info("finish")
}
