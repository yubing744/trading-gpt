package exchange

import (
	"fmt"
	"time"

	"github.com/yubing744/trading-gpt/pkg/types"
)

// EventType constants
const (
	EventPositionClosed = "position_closed"
)

// CloseReason constants
const (
	CloseReasonManual      = "Manual"
	CloseReasonTakeProfit  = "TakeProfit"
	CloseReasonStopLoss    = "StopLoss"
	CloseReasonLiquidation = "Liquidation"
)

// PositionClosedEventData contains all the information about a closed position
type PositionClosedEventData struct {
	StrategyID        string      // ID of the strategy that managed this position
	Symbol            string      // Trading pair symbol
	EntryPrice        float64     // Price at which the position was opened
	ExitPrice         float64     // Price at which the position was closed
	Quantity          float64     // Position size
	ProfitAndLoss     float64     // Profit or loss amount
	CloseReason       string      // Reason for closing: "TakeProfit", "StopLoss", "Manual", "Liquidation", etc.
	Timestamp         time.Time   // Time when the position was closed
	RelatedMarketData interface{} // Optional market data snapshot around close time
}

// NewPositionClosedEvent creates a new position closed event
func NewPositionClosedEvent(data PositionClosedEventData) *types.Event {
	return types.NewEvent(EventPositionClosed, data)
}

// ToPrompts for PositionClosedEvent returns a prompt string representation of the event
func (data PositionClosedEventData) ToPrompts() []string {
	pnlStr := "loss"
	if data.ProfitAndLoss >= 0 {
		pnlStr = "profit"
	}

	prompt := fmt.Sprintf("Position closed for %s:\n"+
		"Strategy: %s\n"+
		"Symbol: %s\n"+
		"Entry Price: %.2f\n"+
		"Exit Price: %.2f\n"+
		"Quantity: %.6f\n"+
		"%s: %.2f\n"+
		"Close Reason: %s\n"+
		"Close Time: %s",
		data.Symbol,
		data.StrategyID,
		data.Symbol,
		data.EntryPrice,
		data.ExitPrice,
		data.Quantity,
		pnlStr,
		data.ProfitAndLoss,
		data.CloseReason,
		data.Timestamp.Format(time.RFC3339),
	)

	return []string{prompt}
}
