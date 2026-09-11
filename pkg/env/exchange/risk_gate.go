package exchange

import (
	"fmt"
	"sync"
	"time"

	"github.com/c9s/bbgo/pkg/fixedpoint"

	"github.com/yubing744/trading-gpt/pkg/config"
)

// RiskEvalInput is the full set of inputs the pure risk evaluation needs to
// decide whether to allow an open/add order. All monetary values are in the
// quote currency; DailyRealizedPnL is signed (negative == loss).
type RiskEvalInput struct {
	// OrderNotional is the notional of the order being submitted (quote).
	OrderNotional fixedpoint.Value
	// PositionNotional is the notional of the currently-open position (quote, abs).
	PositionNotional fixedpoint.Value
	// Leverage is the effective leverage used for this order.
	Leverage fixedpoint.Value
	// DailyRealizedPnL is today's accumulated realized PnL (quote, signed).
	DailyRealizedPnL fixedpoint.Value
	// KillSwitchLatched is true when the daily kill-switch has already tripped
	// earlier today (latched by the stateful gate). Once latched it stays deny
	// until the trading day resets.
	KillSwitchLatched bool
}

// RiskDecision is the result of a pure risk evaluation.
type RiskDecision struct {
	Allow  bool
	Code   string // machine-readable reason code ("" when allowed)
	Reason string // human-readable explanation
}

const (
	RiskCodeKillSwitch  = "daily_loss_kill_switch"
	RiskCodeMaxLeverage = "max_leverage_exceeded"
	RiskCodeMaxOrder    = "max_order_notional_exceeded"
	RiskCodeMaxPosition = "max_position_notional_exceeded"
)

// dailyLossTripped reports whether today's realized loss has reached the
// configured MaxDailyLoss threshold. Loss magnitude is the negation of a
// negative PnL; a zero/positive PnL is never a loss.
func dailyLossTripped(dailyPnL, maxDailyLoss fixedpoint.Value) bool {
	if maxDailyLoss.Sign() <= 0 { // check disabled
		return false
	}
	if dailyPnL.Sign() >= 0 { // no loss
		return false
	}
	loss := dailyPnL.Neg()
	return loss.Compare(maxDailyLoss) >= 0
}

// EvaluateRisk is a pure function: given the current exposure/leverage, today's
// realized PnL and the risk config, it returns an allow/deny decision with a
// reason. It performs no I/O and never triggers any order — a deny simply means
// the caller must not submit the order. Checks are evaluated most-protective
// first (kill-switch), then leverage, single-order notional, total notional.
func EvaluateRisk(in RiskEvalInput, cfg config.RiskControlConfig) RiskDecision {
	if !cfg.IsEnabled() {
		return RiskDecision{Allow: true}
	}

	// 1) Daily-loss kill-switch: latched, or freshly tripped by today's PnL.
	if in.KillSwitchLatched || dailyLossTripped(in.DailyRealizedPnL, cfg.MaxDailyLoss) {
		loss := fixedpoint.Zero
		if in.DailyRealizedPnL.Sign() < 0 {
			loss = in.DailyRealizedPnL.Neg()
		}
		return RiskDecision{
			Allow: false,
			Code:  RiskCodeKillSwitch,
			Reason: fmt.Sprintf(
				"daily loss kill-switch active: realized daily loss %.4f >= limit %.4f",
				loss.Float64(), cfg.MaxDailyLoss.Float64()),
		}
	}

	// 2) Leverage cap.
	if cfg.MaxLeverage.Sign() > 0 && in.Leverage.Compare(cfg.MaxLeverage) > 0 {
		return RiskDecision{
			Allow: false,
			Code:  RiskCodeMaxLeverage,
			Reason: fmt.Sprintf(
				"effective leverage %.4f exceeds max_leverage %.4f",
				in.Leverage.Float64(), cfg.MaxLeverage.Float64()),
		}
	}

	// 3) Single-order notional cap.
	if cfg.MaxOrderQuote.Sign() > 0 && in.OrderNotional.Compare(cfg.MaxOrderQuote) > 0 {
		return RiskDecision{
			Allow: false,
			Code:  RiskCodeMaxOrder,
			Reason: fmt.Sprintf(
				"order notional %.4f exceeds max_order_quote %.4f",
				in.OrderNotional.Float64(), cfg.MaxOrderQuote.Float64()),
		}
	}

	// 4) Total (resulting) position notional cap.
	if cfg.MaxPositionQuote.Sign() > 0 {
		total := in.PositionNotional.Add(in.OrderNotional)
		if total.Compare(cfg.MaxPositionQuote) > 0 {
			return RiskDecision{
				Allow: false,
				Code:  RiskCodeMaxPosition,
				Reason: fmt.Sprintf(
					"resulting position notional %.4f (existing %.4f + order %.4f) exceeds max_position_quote %.4f",
					total.Float64(), in.PositionNotional.Float64(), in.OrderNotional.Float64(),
					cfg.MaxPositionQuote.Float64()),
			}
		}
	}

	return RiskDecision{Allow: true}
}

// RiskGate is the stateful wrapper around EvaluateRisk. It tracks today's
// realized PnL and latches the daily kill-switch, resetting both at each new
// trading day. It is safe for concurrent use.
type RiskGate struct {
	cfg config.RiskControlConfig

	mu            sync.Mutex
	day           time.Time // truncated to the day currently being tracked
	dailyRealized fixedpoint.Value
	killLatched   bool

	// now is injectable for deterministic tests; defaults to time.Now (UTC).
	now func() time.Time
}

// NewRiskGate creates a RiskGate for the given config.
func NewRiskGate(cfg config.RiskControlConfig) *RiskGate {
	g := &RiskGate{
		cfg: cfg,
		now: func() time.Time { return time.Now().UTC() },
	}
	g.day = truncateDay(g.now())
	return g
}

func truncateDay(t time.Time) time.Time {
	t = t.UTC()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

// rolloverLocked resets the daily state when the trading day has changed.
// Caller must hold g.mu.
func (g *RiskGate) rolloverLocked() {
	today := truncateDay(g.now())
	if !today.Equal(g.day) {
		g.day = today
		g.dailyRealized = fixedpoint.Zero
		g.killLatched = false
	}
}

// RecordRealizedPnL accumulates a realized PnL (signed; negative == loss) for
// the current trading day and latches the kill-switch if the daily loss limit
// is reached. This is invoked when a position is closed. It never submits any
// order.
func (g *RiskGate) RecordRealizedPnL(pnl fixedpoint.Value) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.rolloverLocked()
	g.dailyRealized = g.dailyRealized.Add(pnl)

	if !g.killLatched && dailyLossTripped(g.dailyRealized, g.cfg.MaxDailyLoss) {
		g.killLatched = true
	}
}

// Evaluate applies the hard risk gate to a prospective open/add order.
// orderNotional and positionNotional are in quote currency; leverage is the
// effective leverage for the order. It is a defensive check only.
func (g *RiskGate) Evaluate(orderNotional, positionNotional, leverage fixedpoint.Value) RiskDecision {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.rolloverLocked()

	// Refresh the latch in case the threshold is already exceeded.
	if !g.killLatched && dailyLossTripped(g.dailyRealized, g.cfg.MaxDailyLoss) {
		g.killLatched = true
	}

	return EvaluateRisk(RiskEvalInput{
		OrderNotional:     orderNotional,
		PositionNotional:  positionNotional,
		Leverage:          leverage,
		DailyRealizedPnL:  g.dailyRealized,
		KillSwitchLatched: g.killLatched,
	}, g.cfg)
}

// DailyRealizedPnL returns today's accumulated realized PnL (after rollover).
func (g *RiskGate) DailyRealizedPnL() fixedpoint.Value {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.rolloverLocked()
	return g.dailyRealized
}

// KillSwitchActive reports whether the daily kill-switch is currently latched.
func (g *RiskGate) KillSwitchActive() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.rolloverLocked()
	return g.killLatched
}
