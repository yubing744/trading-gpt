package env

import (
	"context"

	"github.com/yubing744/trading-gpt/pkg/types"
)

type IEnvironment interface {
	Actions() []*types.ActionDesc
	OnEvent(cb types.EventCallback)
	SendCommand(ctx context.Context, name string, cmd string, args []string) error
}
