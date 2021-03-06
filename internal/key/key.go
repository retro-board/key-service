package key

import (
	"crypto/rand"
	"math/big"
	"time"

	"github.com/retro-board/key-service/internal/config"
)

type Key struct {
	Config *config.Config
}

type ServiceKey struct {
	Key string
}

type UserKey struct {
	ID      string
	Created time.Time

	UserService    ServiceKey
	RetroService   ServiceKey
	TimerService   ServiceKey
	CompanyService ServiceKey
	BillingService ServiceKey
}

func NewKey(config *config.Config) *Key {
	return &Key{
		Config: config,
	}
}

func (k *Key) GenerateServiceKey(n int) (string, error) {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	b := make([]rune, n)
	for i := range b {
		j, err := rand.Int(rand.Reader, big.NewInt(int64(len(letterRunes))))
		if err != nil {
			return "", err
		}

		b[i] = letterRunes[j.Int64()]
	}
	return string(b), nil
}

func (k *Key) GetKeys(n int) (*ResponseItem, error) {
	userKey, err := k.GenerateServiceKey(n)
	if err != nil {
		return nil, err
	}
	retroKey, err := k.GenerateServiceKey(n)
	if err != nil {
		return nil, err
	}
	timerKey, err := k.GenerateServiceKey(n)
	if err != nil {
		return nil, err
	}
	companyKey, err := k.GenerateServiceKey(n)
	if err != nil {
		return nil, err
	}
	billingKey, err := k.GenerateServiceKey(n)
	if err != nil {
		return nil, err
	}
	permissionKey, err := k.GenerateServiceKey(n)
	if err != nil {
		return nil, err
	}

	return &ResponseItem{
		Status:      "ok",
		User:        userKey,
		Retro:       retroKey,
		Timer:       timerKey,
		Company:     companyKey,
		Billing:     billingKey,
		Permissions: permissionKey,
	}, nil
}

func (k *Key) ValidateServiceKey(key string) bool {
	return k.Config.Local.OnePasswordKey == key
}
