package agent

import (
	"context"

	"github.com/yubing744/trading-bot/pkg/types"
)

type GenResult struct {
	Actions []*types.Action
	Texts   []string
}

type Agent interface {
	SetName(name string)
	SetBackgroup(backgroup string)
	GenActions(ctx context.Context, sessionID string, event *types.Event) (*GenResult, error)
	RegisterActions(ctx context.Context, actions []*types.ActionDesc)
}
