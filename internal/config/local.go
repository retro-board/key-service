package config

import (
	"fmt"

	"github.com/caarlos0/env/v6"

	bugLog "github.com/bugfixes/go-bugfixes/logs"
)

type UserService struct {
	Key     string `env:"USER_SERVICE_KEY"`
	Address string `env:"USER_SERVICE_ADDRESS" envDefault:"https://api.retro-board.it/v1/user"`
}
type CompanyService struct {
	Key     string `env:"COMPANY_SERVICE_KEY"`
	Address string `env:"COMPANY_SERVICE_ADDRESS" envDefault:"https://api.retro-board.it/v1/company"`
}
type TimerService struct {
	Key     string `env:"TIMER_SERVICE_KEY"`
	Address string `env:"TIMER_SERVICE_ADDRESS" envDefault:"https://api.retro-board.it/v1/key"`
}
type RetroService struct {
	Key     string `env:"RETRO_SERVICE_KEY"`
	Address string `env:"RETRO_SERVICE_ADDRESS" envDefault:"https://api.retro-board.it/v1/retro"`
}
type BillingService struct {
	Key     string `env:"BILLING_SERVICE_KEY"`
	Address string `env:"BILLING_SERVICE_ADDRESS" envDefault:"https://api.retro-board.it/v1/billing"`
}

type Services struct {
	UserService
	CompanyService
	TimerService
	RetroService
	BillingService
}

type Local struct {
	KeepLocal   bool `env:"LOCAL_ONLY" envDefault:"false"`
	Development bool `env:"DEVELOPMENT" envDefault:"false"`
	Port        int  `env:"PORT" envDefault:"3000"`

	OnePasswordKey  string `env:"ONE_PASSWORD_KEY"`
	OnePasswordPath string `env:"ONE_PASSWORD_PATH"`

	Services
}

func buildLocal(cfg *Config) error {
	local := &Local{}
	if err := env.Parse(local); err != nil {
		return err
	}
	cfg.Local = *local

	if err := buildServiceKeys(cfg); err != nil {
		return bugLog.Errorf("failed to build service keys: %s", err.Error())
	}
	if err := buildServiceKey(cfg); err != nil {
		return bugLog.Errorf("failed to build service key: %s", err.Error())
	}

	return nil
}

func buildServiceKey(cfg *Config) error {
	onePasswordKeyData, err := cfg.getVaultSecrets(cfg.Local.OnePasswordPath)
	if err != nil {
		return err
	}
	secrets, err := ParseKVSecrets(onePasswordKeyData)
	if err != nil {
		return err
	}
	for _, secret := range secrets {
		switch secret.Key {
		case "password":
			cfg.Local.OnePasswordKey = secret.Value
		}
	}
	return nil
}

func buildServiceKeys(cfg *Config) error {
	vaultSecrets, err := cfg.getVaultSecrets("kv/data/retro-board/api-keys")
	if err != nil {
		return err
	}

	if vaultSecrets == nil {
		return fmt.Errorf("api keys not found in vault")
	}

	secrets, err := ParseKVSecrets(vaultSecrets)
	if err != nil {
		return err
	}

	for _, secret := range secrets {
		switch secret.Key {
		case "retro":
			cfg.Local.Services.RetroService.Key = secret.Value
		case "user":
			cfg.Local.Services.UserService.Key = secret.Value
		case "company":
			cfg.Local.Services.CompanyService.Key = secret.Value
		case "key":
			cfg.Local.Services.TimerService.Key = secret.Value
		case "billing":
			cfg.Local.Services.BillingService.Key = secret.Value
		}
	}

	return nil
}
