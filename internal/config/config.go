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

	defaultPort        = 9889
	defaultExpiration  = 30 * time.Minute
	defaultMaintenance = 10 * time.Minute
)

type Config interface {
	Port() int
	Expiration() time.Duration
	Maintenance() time.Duration
}

type EnvConfig struct {
	port        int
	expiration  time.Duration
	maintenance time.Duration
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
