package config

import (
	"fmt"
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

	Frontend      string `env:"FRONTEND_URL" envDefault:"retro-board.it"`
	FrontendProto string `env:"FRONTEND_PROTO" envDefault:"https"`
	JWTSecret     string `env:"JWT_SECRET" envDefault:"retro-board"`
	TokenSeed     string `env:"TOKEN_SEED" envDefault:"retro-board"`

	Services
}

func buildLocal(cfg *Config) error {
	if err := buildLocalKeys(cfg); err != nil {
		return bugLog.Errorf("failed to build local keys: %s", err.Error())
	}

	if err := buildServiceKeys(cfg); err != nil {
		return bugLog.Errorf("failed to build service keys: %s", err.Error())
	}

	return nil
}

func buildLocalKeys(cfg *Config) error {
	vaultSecrets, err := cfg.getVaultSecrets("kv/data/retro-board/local-keys")
	if err != nil {
		return err
	}

	if vaultSecrets == nil {
		return fmt.Errorf("local keys not found in vault")
	}

	secrets, err := ParseKVSecrets(vaultSecrets)
	if err != nil {
		return err
	}

	for _, secret := range secrets {
		switch secret.Key {
		case "jwt-secret":
			cfg.Local.JWTSecret = secret.Value
			break
		case "company-token":
			cfg.Local.TokenSeed = secret.Value
			break
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
			break
		case "user":
			cfg.Local.Services.UserService.Key = secret.Value
			break
		case "company":
			cfg.Local.Services.CompanyService.Key = secret.Value
			break
		case "key":
			cfg.Local.Services.TimerService.Key = secret.Value
			break
		case "billing":
			cfg.Local.Services.BillingService.Key = secret.Value
			break
		}
	}

	return nil
}
