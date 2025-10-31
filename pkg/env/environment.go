package env

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/yubing744/trading-gpt/pkg/config"
	"github.com/yubing744/trading-gpt/pkg/types"
	"github.com/yubing744/trading-gpt/pkg/utils"
)

var log = logrus.WithField("env", "environment")

type Environment struct {
	entites       map[string]IEntity
	callbacks     []types.EventCallback
	includeEvents []string
}

func NewEnvironment(cfg *config.EnvConfig) *Environment {
	return &Environment{
		entites:       make(map[string]IEntity, 0),
		callbacks:     make([]types.EventCallback, 0),
		includeEvents: cfg.IncludeEvents,
	}
}

func (env *Environment) RegisterEntity(entity IEntity) {
	env.entites[entity.GetID()] = entity
}

func (env *Environment) GetEntity(entityID string) IEntity {
	return env.entites[entityID]
}

func (env *Environment) Actions() []*types.ActionDesc {
	actions := make([]*types.ActionDesc, 0)

	for _, ent := range env.entites {
		for _, action := range ent.Actions() {
			actions = append(actions, &types.ActionDesc{
				Name:        fmt.Sprintf("%s.%s", ent.GetID(), action.Name),
				Description: action.Description,
				Args:        action.Args,
			})
		}
	}

	return actions
}

func (env *Environment) SendCommand(ctx context.Context, fullCmd string, args map[string]string) error {
	dotIndex := strings.Index(fullCmd, ".")
	if dotIndex == -1 || strings.Contains(fullCmd[dotIndex+1:], ".") {
		return errors.New("cmd not correct, can not parse entity_id")
	}

	entityName := fullCmd[:dotIndex]
	cmd := fullCmd[dotIndex+1:]

	if entityName == "" || cmd == "" {
		return errors.New("empty entityName or cmd")
	}

	if env.entites == nil {
		return errors.New("entities map is nil")
	}

	entity, ok := env.entites[entityName]
	if !ok {
		log.
			WithField("entityName", entityName).
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
	ch := make(chan types.IEvent)

	for _, entity := range env.entites {
		go func(ent IEntity) {
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

func (env *Environment) run(ctx context.Context, ch chan types.IEvent) error {
	for {
		select {
		case evt := <-ch:
			env.emitEvent(evt)
		case <-ctx.Done():
			log.Info("env context done")
			return nil
		}
	}
}

func (env *Environment) emitEvent(evt types.IEvent) {
	log.WithField("event", evt).Info("env emit event")

	if !utils.Contains(env.includeEvents, evt.GetType()) {
		log.
			WithField("eventType", evt.GetType()).
			WithField("includeEvents", env.includeEvents).
			Info("skip event for include blacklist")
		return
	}

	for _, cb := range env.callbacks {
		cb(evt)
	}
}
