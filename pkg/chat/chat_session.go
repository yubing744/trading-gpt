package chat

import (
	"github.com/yubing744/trading-bot/pkg/agent"
	"github.com/yubing744/trading-bot/pkg/env"
	"github.com/yubing744/trading-bot/pkg/types"
)

type ChatSession struct {
	User    *types.User
	Channel *types.Channel
	Agent   agent.Agent
	Env     *env.Environment
}
