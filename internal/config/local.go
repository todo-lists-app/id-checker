package config

import (
	"github.com/caarlos0/env/v8"
)

// Local is the local config
type Local struct {
	KeepLocal   bool `env:"BUGFIXES_LOCAL_ONLY" envDefault:"false" json:"keep_local,omitempty"`
	Development bool `env:"DEVELOPMENT" envDefault:"false" json:"development,omitempty"`
	HTTPPort    int  `env:"HTTP_PORT" envDefault:"80" json:"port,omitempty"`
	GRPCPort    int  `env:"GRPC_PORT" envDefault:"3000" json:"grpc_port,omitempty"`
}

// BuildLocal builds the local config
func BuildLocal(cfg *Config) error {
	local := &Local{}
	if err := env.Parse(local); err != nil {
		return err
	}
	cfg.Local = *local

	return nil
}
