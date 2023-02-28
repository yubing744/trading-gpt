package keeper

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yubing744/trading-gpt/pkg/agent"
	"github.com/yubing744/trading-gpt/pkg/config"
	"github.com/yubing744/trading-gpt/pkg/types"
)

type baseAgent struct {
	Name string
}

func (agent *baseAgent) Start() error {
	return nil
}

func (agent *baseAgent) Stop() {

}

func (a *baseAgent) GetName() string {
	return a.Name
}

func (agent *baseAgent) GenActions(ctx context.Context, session types.ISession, msgs []*types.Message) (*agent.GenResult, error) {
	return nil, errors.New("gen error")
}

type testAgentA struct {
	baseAgent
}

type testAgentB struct {
	baseAgent
}

type testAgentC struct {
	baseAgent
}

func (a *testAgentC) GenActions(ctx context.Context, session types.ISession, msgs []*types.Message) (*agent.GenResult, error) {
	result := &agent.GenResult{}
	return result, nil
}

func TestNewAgentKeeper(t *testing.T) {
	cfg := &config.AgentKeeperConfig{
		Enabled:   true,
		Leader:    "agent_a",
		Followers: []string{"agent_b", "agent_c"},
	}
	agents := map[string]agent.IAgent{
		"agent_a": &testAgentA{
			baseAgent: baseAgent{
				Name: "a",
			},
		},
		"agent_b": &testAgentB{
			baseAgent: baseAgent{
				Name: "b",
			},
		},
		"agent_c": &testAgentC{
			baseAgent: baseAgent{
				Name: "c",
			},
		},
	}

	keeper := NewAgentKeeper(cfg, agents)
	assert.NotNil(t, keeper)
}

func TestAgentKeeperGenActions(t *testing.T) {
	ctx := context.Background()

	cfg := &config.AgentKeeperConfig{
		Enabled:   true,
		Leader:    "agent_a",
		Followers: []string{"agent_c"},
	}
	agents := map[string]agent.IAgent{
		"agent_a": &testAgentA{
			baseAgent: baseAgent{
				Name: "a",
			},
		},
		"agent_c": &testAgentC{
			baseAgent: baseAgent{
				Name: "c",
			},
		},
	}

	keeper := NewAgentKeeper(cfg, agents)
	assert.NotNil(t, keeper)

	session := types.NewMockSession("session_1")
	result, err := keeper.GenActions(ctx, session, []*types.Message{
		{
			Text: "1.1 1.2 1.3 1.4 1.5 1.6 1.7",
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestAgentKeeperGenActionsAllError(t *testing.T) {
	ctx := context.Background()

	cfg := &config.AgentKeeperConfig{
		Enabled:   true,
		Leader:    "agent_a",
		Followers: []string{"agent_b"},
	}
	agents := map[string]agent.IAgent{
		"agent_a": &testAgentA{
			baseAgent: baseAgent{
				Name: "a",
			},
		},
		"agent_b": &testAgentB{
			baseAgent: baseAgent{
				Name: "b",
			},
		},
	}

	keeper := NewAgentKeeper(cfg, agents)
	assert.NotNil(t, keeper)

	session := types.NewMockSession("session_1")
	result, err := keeper.GenActions(ctx, session, []*types.Message{
		{
			Text: "1.1 1.2 1.3 1.4 1.5 1.6 1.7",
		},
	})

	assert.Error(t, err)
	assert.Nil(t, result)
}
