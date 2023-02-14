package agent

import (
	"github.com/yubing744/trading-bot/pkg/types"
)

type Agent interface {
	GenAction(sessionID string, event *types.Event) (*types.Action, error)
	RegisterActions(actions []types.ActionDesc)
}
