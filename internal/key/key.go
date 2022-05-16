package key

import (
	"context"
	"github.com/retro-board/key-service/internal/config"
)

type Key struct {
	Config *config.Config
	CTX    context.Context
}

func NewKey(config *config.Config) *Key {
	return &Key{}
}
