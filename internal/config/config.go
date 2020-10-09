package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	LOGLEVEL     = "LOG_LEVEL"
	PORT         = "PORT"
	EXPIRATION   = "EXPIRATION"
	MAINTENANCE  = "MAINTENANCE"
	PREALLOCATED = "PREALLOCATED"
	MODE         = "STORAGE_MODE"

	defaultLogLevel     = LogLevelInfo
	defaultPort         = 9889
	defaultExpiration   = 30 * time.Minute
	defaultMaintenance  = 10 * time.Minute
	defaultPreAllocated = 1024 * 1024
	defaultStorageMode  = StorageModePartitionedMap
)

const (
	StorageModeMap                StorageMode = "map"
	StorageModeSyncMap            StorageMode = "sync-map"
	StorageModePartitionedMap     StorageMode = "partitioned-map"
	StorageModePartitionedSyncMap StorageMode = "partitioned-sync-map"
)

const (
	LogLevelDebug   LogLevel = "DEBUG"
	LogLevelInfo    LogLevel = "INFO"
	LogLevelWarning LogLevel = "WARNING"
	LogLevelError   LogLevel = "ERROR"
)

type StorageMode string

type LogLevel string

type TimeSource interface {
	Now() time.Time
}

type systemTime struct{}

func (systemTime) Now() time.Time {
	return time.Now()
}

type Config interface {
	LogLevel() LogLevel
	Port() int
	Expiration() time.Duration
	Maintenance() time.Duration
	Mode() StorageMode
	PreAllocated() int
	TimeSource() TimeSource
}

type EnvConfig struct {
	logLevel    LogLevel
	port        int
	expiration  time.Duration
	maintenance time.Duration
	mode        StorageMode
	preAllocted int
	timeSource  TimeSource
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

	preAllocated, err := getIntOr(PREALLOCATED, defaultPreAllocated)
	if err != nil {
		return nil, err
	}

	return &EnvConfig{
		logLevel:    parseLogLevel(),
		port:        port,
		expiration:  expiration,
		maintenance: maintenance,
		mode:        parseMode(),
		preAllocted: preAllocated,
		timeSource:  systemTime{},
	}, nil
}

func (o *EnvConfig) LogLevel() LogLevel {
	return o.logLevel
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

func (o *EnvConfig) PreAllocated() int {
	return o.preAllocted
}

func (o *EnvConfig) TimeSource() TimeSource {
	return o.timeSource
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

func parseLogLevel() LogLevel {
	level := LogLevel(strings.ToUpper(getStringOr(LOGLEVEL, string(defaultLogLevel))))

	switch level {
	case LogLevelDebug,
		LogLevelInfo,
		LogLevelWarning,
		LogLevelError:
		return level
	default:
		return defaultLogLevel
	}
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
