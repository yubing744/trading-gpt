package integration

import (
	"context"
	"fmt"
	"testing"

	"github.com/c9s/bbgo/pkg/bbgo"
	"github.com/c9s/bbgo/pkg/fixedpoint"
	"github.com/c9s/bbgo/pkg/types"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"

	"github.com/yubing744/trading-gpt/pkg/config"
	"github.com/yubing744/trading-gpt/pkg/env/exchange"
)

func setupTestSession(t *testing.T) (*bbgo.Environment, *bbgo.ExchangeSession) {
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

	return environ, session
}

func TestExchangeEntityOpenPosition(t *testing.T) {
	_, session := setupTestSession(t)

	symbol := "OPUSDT"
	market, ok := session.Market(symbol)
	assert.True(t, ok)

	ticker, err := session.Exchange.QueryTicker(context.Background(), symbol)
	assert.NoError(t, err)

	closePrice := ticker.Last
	assert.True(t, ok)
	fmt.Printf("lastPrice: %s\n", closePrice)

	position := types.NewPositionFromMarket(market)

	entity := exchange.NewExchangeEntity(
		symbol,
		"5s",
		fixedpoint.NewFromInt(3),
		&config.EnvExchangeConfig{
			KlineNum: 20,
		},
		session,
		session.OrderExecutor,
		position,
	)

	assert.NotNil(t, entity)

	err = entity.OpenPosition(context.Background(), types.SideTypeBuy, closePrice)
	assert.NoError(t, err)
}

func TestExchangeEntityClosePosition50Percent(t *testing.T) {
	environ, session := setupTestSession(t)

	symbol := "OPUSDT"
	market, ok := session.Market(symbol)
	assert.True(t, ok)

	ticker, err := session.Exchange.QueryTicker(context.Background(), symbol)
	assert.NoError(t, err)

	closePrice := ticker.Last
	assert.True(t, ok)
	fmt.Printf("lastPrice: %s\n", closePrice)

	position := types.NewPositionFromMarket(market)

	// Set fee rate
	if session.MakerFeeRate.Sign() > 0 || session.TakerFeeRate.Sign() > 0 {
		position.SetExchangeFeeRate(session.ExchangeName, types.ExchangeFee{
			MakerFeeRate: session.MakerFeeRate,
			TakerFeeRate: session.TakerFeeRate,
		})
	}

	// Setup order executor
	orderExecutor := bbgo.NewGeneralOrderExecutor(session, symbol, "bbgo_test", "bbgo_test_1", position)
	orderExecutor.BindEnvironment(environ)
	orderExecutor.Bind()

	entity := exchange.NewExchangeEntity(
		symbol,
		"5s",
		fixedpoint.NewFromInt(3),
		&config.EnvExchangeConfig{
			KlineNum: 20,
		},
		session,
		orderExecutor,
		position,
	)

	assert.NotNil(t, entity)

	err = entity.ClosePosition(context.Background(), fixedpoint.NewFromFloat(0.5), closePrice)
	assert.NoError(t, err)
}
