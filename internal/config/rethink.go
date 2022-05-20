package config

import "github.com/caarlos0/env/v6"

type Rethink struct {
	Address string `env:"RETHINK_ADDRESS" envDefault:"localhost:28015"`
}

func buildRethink(cfg *Config) error {
	r := &Rethink{}
	if err := env.Parse(r); err != nil {
		return err
	}
	cfg.Rethink = *r

	return nil
}
