package exchange

import (
	"github.com/c9s/bbgo/pkg/datatype/floats"
	"github.com/c9s/bbgo/pkg/fixedpoint"
	"github.com/c9s/bbgo/pkg/types"
)

type PositionX struct {
	*types.Position

	Dust           bool
	historyProfits []fixedpoint.Value
}

func NewPositionX(pos *types.Position) *PositionX {
	x := &PositionX{
		Position:       pos,
		historyProfits: make([]fixedpoint.Value, 0),
	}

	pos.OnModify(func(baseQty fixedpoint.Value, quoteQty fixedpoint.Value, price fixedpoint.Value) {
		if pos.IsClosed() {
			x.historyProfits = make([]fixedpoint.Value, 0)
		}
	})

	return x
}

func (pos *PositionX) AddProfit(val fixedpoint.Value) {
	pos.AccumulatedProfit = val
	pos.historyProfits = append(pos.historyProfits, val)
}

func (pos *PositionX) GetProfitValues() floats.Slice {
	values := make(floats.Slice, 0)

	for _, profit := range pos.historyProfits {
		values.Update(profit.Float64())
	}

	return values
}

func (pos *PositionX) GetHoldingPeriod() int {
	return len(pos.historyProfits)
}
