package fng

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yubing744/trading-gpt/pkg/apis/alternative"
	"github.com/yubing744/trading-gpt/pkg/types"
)

var log = logrus.WithField("entity", "fng")

type FearAndGreedEntity struct {
	interval     time.Duration
	delay        time.Duration
	client       *alternative.AlternativeClient
	eventChannel chan types.IEvent // Store event channel for command execution
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

func (entity *FearAndGreedEntity) Actions() []*types.ActionDesc {
	return []*types.ActionDesc{
		{
			Name:        "refresh_index",
			Description: "Manually refresh the current Fear & Greed Index",
			Args:        []types.ArgmentDesc{},
		},
		{
			Name:        "get_historical_index",
			Description: "Get historical Fear & Greed Index data",
			Args: []types.ArgmentDesc{
				{
					Name:        "limit",
					Description: "Number of historical data points to retrieve (default: 7, max: 30)",
				},
			},
		},
	}
}

func (entity *FearAndGreedEntity) HandleCommand(ctx context.Context, cmd string, args map[string]string) error {
	switch cmd {
	case "refresh_index":
		return entity.executeRefreshIndex(ctx)
	case "get_historical_index":
		return entity.executeGetHistoricalIndex(ctx, args)
	default:
		return fmt.Errorf("unknown command: %s", cmd)
	}
}

// executeRefreshIndex manually refreshes the current Fear & Greed Index
func (entity *FearAndGreedEntity) executeRefreshIndex(ctx context.Context) error {
	if entity.eventChannel == nil {
		return fmt.Errorf("event channel not initialized, command can only be executed during Run()")
	}

	log.Info("Executing refresh_index command")

	index, err := entity.client.GetFearAndGreedIndex(1)
	if err != nil {
		return fmt.Errorf("failed to get Fear & Greed Index: %w", err)
	}

	if index != nil && len(index.Data) > 0 {
		fng := index.Data[0].Value
		entity.eventChannel <- types.NewEvent("fng_changed", &fng)
		log.WithField("value", fng).Info("Fear & Greed Index refreshed successfully")
	} else {
		return fmt.Errorf("no Fear & Greed Index data available")
	}

	return nil
}

// executeGetHistoricalIndex retrieves historical Fear & Greed Index data
func (entity *FearAndGreedEntity) executeGetHistoricalIndex(ctx context.Context, args map[string]string) error {
	if entity.eventChannel == nil {
		return fmt.Errorf("event channel not initialized, command can only be executed during Run()")
	}

	// Parse limit parameter (default: 7, max: 30)
	limit := 7
	if limitStr, ok := args["limit"]; ok && limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil {
			return fmt.Errorf("invalid limit parameter: %s", limitStr)
		}
		if parsedLimit > 30 {
			parsedLimit = 30
		}
		if parsedLimit < 1 {
			parsedLimit = 1
		}
		limit = parsedLimit
	}

	log.WithField("limit", limit).Info("Executing get_historical_index command")

	index, err := entity.client.GetFearAndGreedIndex(limit)
	if err != nil {
		return fmt.Errorf("failed to get historical Fear & Greed Index: %w", err)
	}

	if index != nil && len(index.Data) > 0 {
		// Format historical data as a string
		historicalData := fmt.Sprintf("Historical Fear & Greed Index (last %d days):\n", limit)
		for i, data := range index.Data {
			historicalData += fmt.Sprintf("%d. %s: %s (%s)\n", i+1, data.Timestamp, data.Value, data.ValueClassification)
		}

		entity.eventChannel <- types.NewEvent("fng_historical_data", &historicalData)
		log.WithField("count", len(index.Data)).Info("Historical Fear & Greed Index retrieved successfully")
	} else {
		return fmt.Errorf("no historical Fear & Greed Index data available")
	}

	return nil
}

func (entity *FearAndGreedEntity) Run(ctx context.Context, ch chan types.IEvent) {
	// Store event channel for command execution
	entity.eventChannel = ch

	timer := time.NewTimer(entity.delay)
	ticker := time.NewTicker(entity.interval)

	for {
		select {
		case <-ctx.Done():
			log.Info("fng entity done")
			return
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

func (entity *FearAndGreedEntity) updateIndex(ctx context.Context, ch chan types.IEvent) error {
	index, err := entity.client.GetFearAndGreedIndex(1)
	if err != nil {
		return err
	}

	log.WithField("index", index).Debug("update fng index")

	if index != nil && len(index.Data) > 0 {
		fng := index.Data[0].Value

		ch <- types.NewEvent("fng_changed", &fng)
	}

	return nil
}
