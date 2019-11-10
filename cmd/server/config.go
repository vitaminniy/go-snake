package main

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type config struct {
	HTTP struct {
		Addr            string        `envconfig:"HTTP_ADDR" default:":9090"`
		ReadTimeout     time.Duration `envconfig:"HTTP_READ_TIMEOUT" default:"5s"`
		WriteTimeout    time.Duration `envconfig:"HTTP_WRITE_TIMEOUT" default:"5s"`
		ShutdownTimeout time.Duration `envconfig:"HTTP_SHUTDOWN_TIMEOUT" default:"5s"`
	}
}

func loadConfig() (cfg config, err error) {
	if err := envconfig.Process("", &cfg); err != nil {
		_ = envconfig.Usage("", &cfg)
		return cfg, fmt.Errorf("couldn't load config: %w", err)
	}
	return cfg, nil
}
