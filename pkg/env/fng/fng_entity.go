package fng

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yubing744/trading-bot/pkg/apis/alternative"
	"github.com/yubing744/trading-bot/pkg/types"
)

var log = logrus.WithField("entity", "fng")

type FearAndGreedEntity struct {
	interval time.Duration
	delay    time.Duration
	client   *alternative.AlternativeClient
}

func NewFearAndGreedEntity() *FearAndGreedEntity {
	return &FearAndGreedEntity{
		interval: time.Hour,
		delay:    time.Minute,
		client:   alternative.NewAlternativeClient(alternative.WithTimeout(time.Second * 20)),
	}
}

func (entity *FearAndGreedEntity) GetID() string {
	return "fng"
}

func (entity *FearAndGreedEntity) HandleCommand(ctx context.Context, cmd string, args []string) error {
	return nil
}

func (entity *FearAndGreedEntity) Run(ctx context.Context, ch chan *types.Event) {
	timer := time.NewTimer(entity.delay)
	ticker := time.NewTicker(entity.interval)

	for {
		select {
		case <-ctx.Done():
			log.Info("fng entity done")
			break
		case <-timer.C:
			err := entity.updateIndex(ctx, ch)
			if err != nil {
				log.WithError(err).Error("delay update index error")
			}
		case <-ticker.C:
			err := entity.updateIndex(ctx, ch)
			if err != nil {
				log.WithError(err).Error("ticker update index error")
			}
		}
	}
}

func (entity *FearAndGreedEntity) updateIndex(ctx context.Context, ch chan *types.Event) error {
	index, err := entity.client.GetFearAndGreedIndex(1)
	if err != nil {
		return err
	}

	log.WithField("index", index).Debug("update fng index")

	if index != nil && len(index.Data) > 0 {
		fng := index.Data[0].Value

		ch <- &types.Event{
			Type: "fng_changed",
			Data: &fng,
		}
	}

	return nil
}
