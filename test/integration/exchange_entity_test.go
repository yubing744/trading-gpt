package integration

import (
	"context"
	"testing"

	"github.com/c9s/bbgo/pkg/bbgo"
	"github.com/c9s/bbgo/pkg/types"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"

	"github.com/yubing744/trading-gpt/pkg/config"
	"github.com/yubing744/trading-gpt/pkg/env/exchange"
)

func setupTestSession(t *testing.T) *bbgo.ExchangeSession {
	ctx := context.Background()

	err := godotenv.Load("../../.env.local")
	assert.NoError(t, err)

	// load successfully
	userConfig, err := bbgo.Load("../../bbgo.yaml", false)
	assert.NoError(t, err)

	environ := bbgo.NewEnvironment()
	err = environ.ConfigureExchangeSessions(userConfig)
	assert.NoError(t, err)

	err = environ.Init(ctx)
	assert.NoError(t, err)

	session, ok := environ.Session("okex")
	assert.True(t, ok)

	assert.NotNil(t, session)

	return session
}

func TestExchangeEntityOpenPosition(t *testing.T) {
	session := setupTestSession(t)

	symbol := "OPUSDT"
	market, ok := session.Market(symbol)
	assert.True(t, ok)

	position := types.NewPositionFromMarket(market)

	entity := exchange.NewExchangeEntity(
		symbol,
		"5s",
		3,
		&config.EnvExchangeConfig{
			WindowSize: 20,
		},
		session,
		session.OrderExecutor,
		position,
	)

	assert.NotNil(t, entity)

	err := entity.OpenPosition(context.Background(), "buy", 1)
	assert.NoError(t, err)
}
