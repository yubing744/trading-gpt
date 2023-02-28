package keeper

import (
	"context"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/yubing744/trading-gpt/pkg/agent"
	"github.com/yubing744/trading-gpt/pkg/config"
	"github.com/yubing744/trading-gpt/pkg/types"
)

var log = logrus.WithField("agent", "keeper")

type AgentKeeper struct {
	leader    agent.IAgent
	followers []agent.IAgent
}

func findAgents(agents map[string]agent.IAgent, leaderName string, followerNames []string) (agent.IAgent, []agent.IAgent, error) {
	leader, ok := agents[leaderName]
	if !ok {
		return nil, nil, errors.Errorf("not found agent by name: %s", leaderName)
	}

	followers := make([]agent.IAgent, 0)
	for _, followerName := range followerNames {
		follower, ok := agents[followerName]
		if !ok {
			return nil, nil, errors.Errorf("not found agent by name: %s", followerName)
		}

		followers = append(followers, follower)
	}

	return leader, followers, nil
}

func NewAgentKeeper(cfg *config.AgentKeeperConfig, agents map[string]agent.IAgent) *AgentKeeper {
	leader, followers, err := findAgents(agents, cfg.Leader, cfg.Followers)
	if err != nil {
		log.WithError(err).Fatal("init keeper error")
	}

	return &AgentKeeper{
		leader:    leader,
		followers: followers,
	}
}

func (agent *AgentKeeper) Start() error {
	if agent.leader != nil {
		err := agent.leader.Start()
		if err != nil {
			return errors.Wrap(err, "Error in start leader agent")
		}
	}

	for _, follower := range agent.followers {
		err := follower.Start()
		if err != nil {
			return errors.Wrap(err, "Error in start follower agent")
		}
	}

	return nil
}

func (agent *AgentKeeper) Stop() {
	if agent.leader != nil {
		agent.leader.Stop()
	}

	for _, follower := range agent.followers {
		follower.Stop()
	}
}

func (a *AgentKeeper) GetName() string {
	return "keeper"
}

func (agent *AgentKeeper) GenActions(ctx context.Context, session types.ISession, msgs []*types.Message) (*agent.GenResult, error) {
	result, err := agent.leader.GenActions(ctx, session, msgs)
	if err == nil {
		return result, nil
	}

	log.WithError(err).Error("gen actions error")

	for _, follower := range agent.followers {
		log.Infof("try follower %s ...", follower.GetName())

		result, ferr := follower.GenActions(ctx, session, msgs)
		if ferr == nil {
			return result, nil
		}

		log.WithError(err).Errorf("follower %s gen actions error", follower.GetName())
	}

	return nil, err
}
