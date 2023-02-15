package agent

import (
	"context"

	"github.com/yubing744/trading-bot/pkg/types"
)

type Agent interface {
	SetName(name string)
	SetBackgroup(backgroup string)
	GenActions(ctx context.Context, sessionID string, event *types.Event) ([]*types.Action, error)
	RegisterActions(ctx context.Context, actions []*types.ActionDesc)
}
