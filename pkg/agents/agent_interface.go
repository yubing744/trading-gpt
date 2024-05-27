package agents

import (
	"context"

	"github.com/yubing744/trading-gpt/pkg/types"
)

type GenResult struct {
	Texts []string
	Model string
}

type IAgent interface {
	types.ICompoment

	GetName() string
	GenActions(ctx context.Context, session types.ISession, msgs []*types.Message) (*GenResult, error)
}
