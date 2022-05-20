package key

import (
	"context"
	"github.com/retro-board/key-service/internal/config"
	"math/rand"
	"time"
)

type Key struct {
	Config *config.Config
	CTX    context.Context
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
	return &Key{}
}

func (k *Key) generateServiceKey(n int) string {
	rand.Seed(time.Now().UnixNano())
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
