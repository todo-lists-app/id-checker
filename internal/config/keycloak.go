package config

import "github.com/caarlos0/env/v8"

type Keycloak struct {
	Client string `env:"KEYCLOAK_CLIENT" envDefault:"" json:"client,omitempty"`
	Secret string `env:"KEYCLOAK_SECRET" envDefault:"" json:"secret,omitempty"`
	Realm  string `env:"KEYCLOAK_REALM" envDefault:"" json:"realm,omitempty"`
	Host   string `env:"KEYCLOAK_HOSTNAME" envDefault:"" json:"host,omitempty"`
}

func BuildKeyCloak(cfg *Config) error {
	keycloak := &Keycloak{}
	if err := env.Parse(keycloak); err != nil {
		return err
	}
	cfg.Keycloak = *keycloak

	return nil
}
