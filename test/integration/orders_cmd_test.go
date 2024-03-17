package integration

import (
	"context"
	"testing"

	"github.com/yubing744/trading-gpt/cmd"
)

func TestListOpenOrders(t *testing.T) {
	ctx := context.Background()

	cmd.Execute(ctx, []string{
		"--dotenv=../../.env.local",
		"--config=../../bbgo.yaml",
		"list-orders",
		"open",
		"--session=okex",
		"--symbol=OPUSDT",
	})
}

func TestSubmitOrders(t *testing.T) {
	ctx := context.Background()

	// submit-order --session=ftx --symbol=BTCUSDT --side=buy --price=18000 --quantity=0.001
	cmd.Execute(ctx, []string{
		"--dotenv=../../.env.local",
		"--config=../../bbgo.yaml",
		"submit-order",
		"--session=okex",
		"--symbol=OPUSDT",
		"--side=buy",
		"--market=true",
		"--quantity=1",
	})
}
