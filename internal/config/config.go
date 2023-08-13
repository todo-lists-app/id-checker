// Package config is used to build the configuration for the service.
package config

import (
	"github.com/bugfixes/go-bugfixes/logs"
	"github.com/caarlos0/env/v8"
)

// Config is the main config
type Config struct {
	Keycloak
	Local
}

// Build is used to build the config, it will call BuildVault and BuildMongo
func Build() (*Config, error) {
	cfg := &Config{}

	if err := env.Parse(cfg); err != nil {
		return nil, logs.Error(err)
	}

	if err := BuildKeyCloak(cfg); err != nil {
		return nil, logs.Error(err)
	}

	if err := BuildLocal(cfg); err != nil {
		return nil, logs.Error(err)
	}

	return cfg, nil
}
