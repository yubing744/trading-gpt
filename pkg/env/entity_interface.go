package env

import (
	"context"

	"github.com/yubing744/trading-gpt/pkg/types"
)

type IEntity interface {
	GetID() string
	Actions() []*types.ActionDesc
	HandleCommand(ctx context.Context, cmd string, args map[string]string) error
	Run(ctx context.Context, ch chan types.IEvent)
}
