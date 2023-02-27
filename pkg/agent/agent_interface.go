package agent

import (
	"context"

	"github.com/yubing744/trading-bot/pkg/types"
)

type GenResult struct {
	Texts []string
}

type IAgent interface {
	types.ICompoment

	GetName() string
	GenActions(ctx context.Context, session types.ISession, msgs []*types.Message) (*GenResult, error)
}
