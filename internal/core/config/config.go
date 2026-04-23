package core_config

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	TimeZone *time.Location
}

type envConfig struct {
	TimeZone string `envconfig:"TIME_ZONE" default:"UTC"`
}

func NewConfig() (Config, error) {
	var env envConfig

	if err := envconfig.Process("", &env); err != nil {
		return Config{}, fmt.Errorf("process envconfig: %w", err)
	}

	zone, err := time.LoadLocation(env.TimeZone)
	if err != nil {
		return Config{}, fmt.Errorf("load time zone: %s: %w", env.TimeZone, err)
	}

	return Config{
		TimeZone: zone,
	}, nil
}

func NewConfigMust() Config {
	config, err := NewConfig()
	if err != nil {
		err = fmt.Errorf("get core config: %w", err)
		panic(err)
	}

	return config
}
