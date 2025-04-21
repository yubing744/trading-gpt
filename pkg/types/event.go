package types

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// EventType constants
const (
	EventPositionClosed = "position_closed"
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

type IEvent interface {
	GetID() string
	GetType() string
	GetData() interface{}
	ToPrompts() []string
}

type Event struct {
	id    string
	ttype string
	data  interface{}
}

func NewEvent(ty string, data interface{}) *Event {
	return &Event{
		id:    uuid.NewString(),
		ttype: ty,
		data:  data,
	}
}

func (e *Event) GetID() string {
	return e.id
}

func (e *Event) GetType() string {
	return e.ttype
}

func (e *Event) GetData() interface{} {
	return e.data
}

func (e *Event) ToPrompts() []string {
	return []string{}
}

// NewPositionClosedEvent creates a new position closed event
func NewPositionClosedEvent(data PositionClosedEventData) *Event {
	return NewEvent(EventPositionClosed, data)
}

// ToPrompts for PositionClosedEvent returns a prompt string representation of the event
func (e *Event) ToPositionClosedPrompts() []string {
	if e.ttype != EventPositionClosed {
		return []string{}
	}

	data, ok := e.data.(PositionClosedEventData)
	if !ok {
		return []string{}
	}

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
