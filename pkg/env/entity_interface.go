package env

import (
	"context"

	"github.com/yubing744/trading-bot/pkg/types"
)

type Entity interface {
	GetID() string
	HandleCommand(ctx context.Context, cmd string, args []string)
	Run(ctx context.Context, ch chan *types.Event)
}
