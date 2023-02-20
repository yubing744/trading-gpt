package agent

import (
	"context"

	"github.com/yubing744/trading-bot/pkg/types"
)

type GenResult struct {
	Actions []*types.Action
	Texts   []string
}

type IAgent interface {
	SetName(name string)
	SetBackgroup(backgroup string)
	RegisterActions(ctx context.Context, name string, actions []*types.ActionDesc)
	Init() error
	GenActions(ctx context.Context, session types.ISession, msgs []*types.Message) (*GenResult, error)
}
