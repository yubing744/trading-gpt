package exchange

import (
	"testing"
	"time"

	"github.com/c9s/bbgo/pkg/fixedpoint"

	"github.com/yubing744/trading-gpt/pkg/config"
)

func fp(f float64) fixedpoint.Value { return fixedpoint.NewFromFloat(f) }

func boolPtr(b bool) *bool { return &b }

// baseCfg is a fully-configured (all limits set) gate config used by most tests.
func baseCfg() config.RiskControlConfig {
	return config.RiskControlConfig{
		Enabled:          boolPtr(true),
		MaxLeverage:      fp(3),
		MaxOrderQuote:    fp(100),
		MaxPositionQuote: fp(150),
		MaxDailyLoss:     fp(50),
	}
}

func TestEvaluateRisk_AllowsWithinAllLimits(t *testing.T) {
	in := RiskEvalInput{
		OrderNotional:    fp(45),
		PositionNotional: fp(0),
		Leverage:         fp(3),
		DailyRealizedPnL: fp(-10),
	}
	d := EvaluateRisk(in, baseCfg())
	if !d.Allow {
		t.Fatalf("expected allow, got deny: %s", d.Reason)
	}
	if d.Code != "" {
		t.Fatalf("expected empty code on allow, got %q", d.Code)
	}
}

func TestEvaluateRisk_DisabledMasterSwitchAllowsEverything(t *testing.T) {
	cfg := baseCfg()
	cfg.Enabled = boolPtr(false)
	in := RiskEvalInput{
		OrderNotional:     fp(10000),
		PositionNotional:  fp(10000),
		Leverage:          fp(100),
		DailyRealizedPnL:  fp(-9999),
		KillSwitchLatched: true,
	}
	if d := EvaluateRisk(in, cfg); !d.Allow {
		t.Fatalf("disabled gate must allow, got deny: %s", d.Reason)
	}
}

func TestEvaluateRisk_ZeroThresholdsDisableIndividualChecks(t *testing.T) {
	// Enabled but no thresholds set -> everything allowed (no behaviour change).
	cfg := config.RiskControlConfig{Enabled: boolPtr(true)}
	in := RiskEvalInput{
		OrderNotional:    fp(1e9),
		PositionNotional: fp(1e9),
		Leverage:         fp(125),
		DailyRealizedPnL: fp(-1e9),
	}
	if d := EvaluateRisk(in, cfg); !d.Allow {
		t.Fatalf("zero thresholds must allow, got deny: %s", d.Reason)
	}
}

func TestEvaluateRisk_MaxLeverage(t *testing.T) {
	cfg := baseCfg()
	// exactly at limit -> allowed
	if d := EvaluateRisk(RiskEvalInput{Leverage: fp(3), OrderNotional: fp(1)}, cfg); !d.Allow {
		t.Fatalf("leverage at limit must be allowed")
	}
	// just above -> denied
	d := EvaluateRisk(RiskEvalInput{Leverage: fp(3.0001), OrderNotional: fp(1)}, cfg)
	if d.Allow || d.Code != RiskCodeMaxLeverage {
		t.Fatalf("expected max_leverage deny, got allow=%v code=%s", d.Allow, d.Code)
	}
}

func TestEvaluateRisk_MaxOrderQuote(t *testing.T) {
	cfg := baseCfg()
	if d := EvaluateRisk(RiskEvalInput{OrderNotional: fp(100), Leverage: fp(1)}, cfg); !d.Allow {
		t.Fatalf("order notional at limit must be allowed")
	}
	d := EvaluateRisk(RiskEvalInput{OrderNotional: fp(100.01), Leverage: fp(1)}, cfg)
	if d.Allow || d.Code != RiskCodeMaxOrder {
		t.Fatalf("expected max_order deny, got allow=%v code=%s", d.Allow, d.Code)
	}
}

func TestEvaluateRisk_MaxPositionQuote(t *testing.T) {
	cfg := baseCfg()
	// existing 120 + order 30 = 150 == limit -> allowed
	if d := EvaluateRisk(RiskEvalInput{PositionNotional: fp(120), OrderNotional: fp(30), Leverage: fp(1)}, cfg); !d.Allow {
		t.Fatalf("total notional at limit must be allowed")
	}
	// existing 120 + order 31 = 151 > 150 -> denied
	d := EvaluateRisk(RiskEvalInput{PositionNotional: fp(120), OrderNotional: fp(31), Leverage: fp(1)}, cfg)
	if d.Allow || d.Code != RiskCodeMaxPosition {
		t.Fatalf("expected max_position deny, got allow=%v code=%s", d.Allow, d.Code)
	}
}

func TestEvaluateRisk_KillSwitchFromDailyLoss(t *testing.T) {
	cfg := baseCfg() // MaxDailyLoss = 50
	// loss of 49 -> still allowed
	if d := EvaluateRisk(RiskEvalInput{DailyRealizedPnL: fp(-49), OrderNotional: fp(1), Leverage: fp(1)}, cfg); !d.Allow {
		t.Fatalf("loss below limit must be allowed")
	}
	// loss of exactly 50 -> tripped
	d := EvaluateRisk(RiskEvalInput{DailyRealizedPnL: fp(-50), OrderNotional: fp(1), Leverage: fp(1)}, cfg)
	if d.Allow || d.Code != RiskCodeKillSwitch {
		t.Fatalf("loss at limit must trip kill-switch, got allow=%v code=%s", d.Allow, d.Code)
	}
	// profit -> not tripped
	if d := EvaluateRisk(RiskEvalInput{DailyRealizedPnL: fp(1000), OrderNotional: fp(1), Leverage: fp(1)}, cfg); !d.Allow {
		t.Fatalf("profit must never trip kill-switch")
	}
}

func TestEvaluateRisk_LatchedKillSwitchDeniesEvenWithoutLoss(t *testing.T) {
	cfg := baseCfg()
	d := EvaluateRisk(RiskEvalInput{
		DailyRealizedPnL:  fp(0),
		KillSwitchLatched: true,
		OrderNotional:     fp(1),
		Leverage:          fp(1),
	}, cfg)
	if d.Allow || d.Code != RiskCodeKillSwitch {
		t.Fatalf("latched kill-switch must deny, got allow=%v code=%s", d.Allow, d.Code)
	}
}

func TestEvaluateRisk_KillSwitchTakesPrecedence(t *testing.T) {
	// Even if leverage/notional are fine, a tripped kill-switch denies first.
	cfg := baseCfg()
	d := EvaluateRisk(RiskEvalInput{
		DailyRealizedPnL: fp(-60),
		OrderNotional:    fp(1),
		Leverage:         fp(1),
	}, cfg)
	if d.Code != RiskCodeKillSwitch {
		t.Fatalf("kill-switch must take precedence, got code=%s", d.Code)
	}
}

// --- Stateful RiskGate tests ---

func newGateAt(cfg config.RiskControlConfig, clock *time.Time) *RiskGate {
	g := NewRiskGate(cfg)
	g.now = func() time.Time { return *clock }
	*clock = time.Date(2026, 7, 13, 12, 0, 0, 0, time.UTC)
	g.day = truncateDay(*clock)
	return g
}

func TestRiskGate_RecordAndTripKillSwitch(t *testing.T) {
	clock := time.Now()
	g := newGateAt(baseCfg(), &clock)

	g.RecordRealizedPnL(fp(-20))
	if g.KillSwitchActive() {
		t.Fatalf("should not be tripped after -20 (< 50)")
	}
	g.RecordRealizedPnL(fp(-35)) // cumulative -55 >= 50
	if !g.KillSwitchActive() {
		t.Fatalf("should be tripped after cumulative -55")
	}

	d := g.Evaluate(fp(10), fp(0), fp(1))
	if d.Allow || d.Code != RiskCodeKillSwitch {
		t.Fatalf("gate must deny while kill-switch latched, got allow=%v code=%s", d.Allow, d.Code)
	}
}

func TestRiskGate_KillSwitchLatchesDespiteRecovery(t *testing.T) {
	clock := time.Now()
	g := newGateAt(baseCfg(), &clock)

	g.RecordRealizedPnL(fp(-60)) // trip
	g.RecordRealizedPnL(fp(100)) // recover to +40 same day
	if !g.KillSwitchActive() {
		t.Fatalf("kill-switch must stay latched for the rest of the day despite recovery")
	}
	if d := g.Evaluate(fp(10), fp(0), fp(1)); d.Allow {
		t.Fatalf("gate must keep denying while latched same day")
	}
}

func TestRiskGate_ResetsOnNewTradingDay(t *testing.T) {
	clock := time.Now()
	g := newGateAt(baseCfg(), &clock)

	g.RecordRealizedPnL(fp(-60)) // trip today
	if !g.KillSwitchActive() {
		t.Fatalf("expected tripped on day 1")
	}

	// advance to next day
	clock = time.Date(2026, 7, 14, 0, 0, 1, 0, time.UTC)

	if g.KillSwitchActive() {
		t.Fatalf("kill-switch must reset on new trading day")
	}
	if g.DailyRealizedPnL().Sign() != 0 {
		t.Fatalf("daily realized PnL must reset to zero on new day, got %v", g.DailyRealizedPnL().Float64())
	}
	if d := g.Evaluate(fp(45), fp(0), fp(3)); !d.Allow {
		t.Fatalf("gate must allow normal order on the new day, got deny: %s", d.Reason)
	}
}

func TestRiskGate_NotionalChecksThroughGate(t *testing.T) {
	clock := time.Now()
	g := newGateAt(baseCfg(), &clock)

	if d := g.Evaluate(fp(45), fp(0), fp(3)); !d.Allow {
		t.Fatalf("normal order must be allowed")
	}
	if d := g.Evaluate(fp(200), fp(0), fp(3)); d.Allow || d.Code != RiskCodeMaxOrder {
		t.Fatalf("oversized single order must be denied, got allow=%v code=%s", d.Allow, d.Code)
	}
	if d := g.Evaluate(fp(40), fp(120), fp(3)); d.Allow || d.Code != RiskCodeMaxPosition {
		t.Fatalf("oversized total must be denied, got allow=%v code=%s", d.Allow, d.Code)
	}
}

func TestRiskGate_DefaultEnabledWhenNil(t *testing.T) {
	cfg := config.RiskControlConfig{MaxDailyLoss: fp(50)} // Enabled nil -> true
	clock := time.Now()
	g := newGateAt(cfg, &clock)
	g.RecordRealizedPnL(fp(-60))
	if d := g.Evaluate(fp(1), fp(0), fp(1)); d.Allow {
		t.Fatalf("nil Enabled must default to enabled; kill-switch should deny")
	}
}

func TestRiskControlConfig_IsEnabled(t *testing.T) {
	if !(config.RiskControlConfig{}).IsEnabled() {
		t.Fatalf("nil Enabled must default to true")
	}
	if (config.RiskControlConfig{Enabled: boolPtr(false)}).IsEnabled() {
		t.Fatalf("explicit false must disable")
	}
}
