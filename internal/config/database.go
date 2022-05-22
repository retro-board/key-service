package config

import (
	"errors"

	"github.com/caarlos0/env/v6"
)

type Database struct {
	Host     string `env:"RDS_HOSTNAME" envDefault:"postgres.retro-board"`
	Port     int    `env:"RDS_PORT" envDefault:"5432"`
	User     string `env:"RDS_USERNAME"`
	Password string `env:"RDS_PASSWORD"`
	DBName   string `env:"RDS_DB" envDefault:"postgres"`
}

func buildDatabase(c *Config) error {
	rds := &Database{}

	if err := env.Parse(rds); err != nil {
		return err
	}

	if rds.User != "" && rds.Password != "" {
		c.Database = *rds
		return nil
	}

	creds, err := c.getVaultSecrets("database/creds/retro-board-backend")
	if err != nil {
		return err
	}

	if creds == nil {
		return errors.New("no database password found")
	}

	rds.Password = creds["password"].(string)
	rds.User = creds["username"].(string)
	c.Database = *rds

	return nil
}
