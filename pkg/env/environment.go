package env

import (
	"context"
	"errors"

	"github.com/sirupsen/logrus"
	"github.com/yubing744/trading-bot/pkg/config"
	"github.com/yubing744/trading-bot/pkg/types"
	"github.com/yubing744/trading-bot/pkg/utils"
)

var log = logrus.WithField("env", "environment")

type Environment struct {
	entites       map[string]Entity
	callbacks     []types.EventCallback
	includeEvents []string
}

func NewEnvironment(cfg *config.EnvConfig) *Environment {
	return &Environment{
		entites:       make(map[string]Entity, 0),
		callbacks:     make([]types.EventCallback, 0),
		includeEvents: cfg.IncludeEvents,
	}
}

func (env *Environment) RegisterEntity(entity Entity) {
	env.entites[entity.GetID()] = entity
}

func (env *Environment) Actions() []*types.ActionDesc {
	actions := make([]*types.ActionDesc, 0)

	for _, ent := range env.entites {
		as := ent.Actions()

		if as != nil {
			actions = append(actions, as...)
		}
	}

	return actions
}

func (env *Environment) SendCommand(ctx context.Context, name string, cmd string, args []string) error {
	entity, ok := env.entites[name]
	if !ok {
		log.
			WithField("entityName", name).
			WithField("cmd", cmd).
			WithField("args", args).
			Debug("not found entity")

		return errors.New("entity not found")
	}

	return entity.HandleCommand(ctx, cmd, args)
}

func (env *Environment) OnEvent(cb types.EventCallback) {
	env.callbacks = append(env.callbacks, cb)
}

func (env *Environment) Start(ctx context.Context) error {
	ch := make(chan *types.Event)

	for _, entity := range env.entites {
		go func(ent Entity) {
			ent.Run(ctx, ch)
		}(entity)
	}

	go func() {
		env.run(ctx, ch)
	}()

	return nil
}

func (env *Environment) Stop(ctx context.Context) {

}

func (env *Environment) run(ctx context.Context, ch chan *types.Event) error {
	for {
		select {
		case evt := <-ch:
			env.emitEvent(evt)
		case <-ctx.Done():
			log.Info("env context done")
			break
		}
	}
}

func (env *Environment) emitEvent(evt *types.Event) {
	log.WithField("event", evt).Info("env emit event")

	if !utils.Contains(env.includeEvents, evt.Type) {
		log.
			WithField("eventType", evt.Type).
			WithField("includeEvents", env.includeEvents).
			Info("skip event for include blacklist")
		return
	}

	for _, cb := range env.callbacks {
		cb(evt)
	}
}
