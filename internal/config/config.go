package config

import (
	"os"
	"strconv"
	"time"
)

const (
	PORT        = "PORT"
	EXPIRATION  = "EXPIRATION"
	MAINTENANCE = "MAINTENANCE"
	MODE        = "STORAGE_MODE"

	defaultPort        = 9889
	defaultExpiration  = 30 * time.Minute
	defaultMaintenance = 10 * time.Minute
	defaultStorageMode = StorageModePartitionedMap
)

const (
	StorageModeMap                StorageMode = "map"
	StorageModeSyncMap            StorageMode = "sync-map"
	StorageModePartitionedMap     StorageMode = "partitioned-map"
	StorageModePartitionedSyncMap StorageMode = "partitioned-sync-map"
)

type StorageMode string

type Config interface {
	Port() int
	Expiration() time.Duration
	Maintenance() time.Duration
	Mode() StorageMode
}

type EnvConfig struct {
	port        int
	expiration  time.Duration
	maintenance time.Duration
	mode        StorageMode
}

func NewConfigFromEnv() (Config, error) {
	port, err := getIntOr(PORT, defaultPort)
	if err != nil {
		return nil, err
	}

	expiration, err := getDurationOr(EXPIRATION, defaultExpiration)
	if err != nil {
		return nil, err
	}

	maintenance, err := getDurationOr(MAINTENANCE, defaultMaintenance)
	if err != nil {
		return nil, err
	}

	return &EnvConfig{
		port:        port,
		expiration:  expiration,
		maintenance: maintenance,
		mode:        parseMode(),
	}, nil
}

func (o *EnvConfig) Port() int {
	return o.port
}

func (o *EnvConfig) Expiration() time.Duration {
	return o.expiration
}

func (o *EnvConfig) Maintenance() time.Duration {
	return o.maintenance
}

func (o *EnvConfig) Mode() StorageMode {
	return o.mode
}

func getIntOr(key string, defV int) (int, error) {
	v := os.Getenv(key)
	if v == "" {
		return defV, nil
	}

	return strconv.Atoi(v)
}

func getDurationOr(key string, defV time.Duration) (time.Duration, error) {
	v := os.Getenv(key)
	if v == "" {
		return defV, nil
	}

	return time.ParseDuration(v)
}

func getStringOr(key string, defV string) string {
	v := os.Getenv(key)
	if v == "" {
		return defV
	}

	return v
}

func parseMode() StorageMode {
	mode := StorageMode(getStringOr(MODE, string(defaultStorageMode)))

	switch mode {
	case StorageModeMap,
		StorageModeSyncMap,
		StorageModePartitionedMap,
		StorageModePartitionedSyncMap:
		return mode
	default:
		return defaultStorageMode
	}
}
