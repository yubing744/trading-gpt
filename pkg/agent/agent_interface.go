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
	RegisterActions(ctx context.Context, name string, actions []*types.ActionDesc)
	GenActions(ctx context.Context, session types.ISession, event *types.Event) (*GenResult, error)
}
