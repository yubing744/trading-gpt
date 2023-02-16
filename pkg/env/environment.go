package env

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/yubing744/trading-bot/pkg/types"
)

var log = logrus.WithField("env", "environment")

type Environment struct {
	entites   map[string]Entity
	callbacks []types.EventCallback
}

func NewEnvironment() *Environment {
	return &Environment{
		entites:   make(map[string]Entity, 0),
		callbacks: make([]types.EventCallback, 0),
	}
}

func (env *Environment) RegisterEntity(entity Entity) {
	env.entites[entity.GetID()] = entity
}

func (env *Environment) SendCommand(ctx context.Context, name string, cmd string, args []string) {
	entity, ok := env.entites[name]
	if ok {
		entity.HandleCommand(ctx, cmd, args)
	} else {
		log.
			WithField("entityName", name).
			WithField("cmd", cmd).
			WithField("args", args).
			Debug("not found entity")
	}
}

func (env *Environment) OnEvent(cb types.EventCallback) {
	env.callbacks = append(env.callbacks, cb)
}

func (env *Environment) Start(ctx context.Context) error {
	ch := make(chan *types.Event, 0)

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
	log.WithField("evnet", evt).Info("env emit event")

	for _, cb := range env.callbacks {
		cb(evt)
	}
}
