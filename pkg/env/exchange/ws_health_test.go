package exchange

import (
	"testing"
	"time"
)

func TestReconnectBackoff(t *testing.T) {
	base := 30 * time.Second
	max := 10 * time.Minute

	cases := []struct {
		attempt int
		want    time.Duration
	}{
		{-1, base},
		{0, base},
		{1, 1 * time.Minute},
		{2, 2 * time.Minute},
		{3, 4 * time.Minute},
		{4, 8 * time.Minute},
		{5, max}, // 16m capped to 10m
		{100, max},
	}

	for _, c := range cases {
		got := ReconnectBackoff(c.attempt, base, max)
		if got != c.want {
			t.Errorf("ReconnectBackoff(%d) = %v, want %v", c.attempt, got, c.want)
		}
	}
}

func TestReconnectBackoffDefaults(t *testing.T) {
	// base<=0 and max<=0 fall back to defaults.
	if got := ReconnectBackoff(0, 0, 0); got != DefaultWSDisconnectAlertBase {
		t.Errorf("ReconnectBackoff with zero base = %v, want %v", got, DefaultWSDisconnectAlertBase)
	}
	if got := ReconnectBackoff(100, 0, 0); got != DefaultWSDisconnectAlertMax {
		t.Errorf("ReconnectBackoff large attempt = %v, want %v", got, DefaultWSDisconnectAlertMax)
	}
}

func TestWSHealthMonitorReconnectStateMachine(t *testing.T) {
	m := NewWSHealthMonitor()
	now := time.Now()

	if m.Connected() {
		t.Fatal("new monitor should not be connected")
	}
	if m.ReconnectCount() != 0 {
		t.Fatalf("new monitor reconnect count = %d, want 0", m.ReconnectCount())
	}

	// First connect is not a reconnect.
	m.RecordConnect(now)
	if !m.Connected() {
		t.Fatal("should be connected after RecordConnect")
	}
	if m.ReconnectCount() != 0 {
		t.Fatalf("first connect reconnect count = %d, want 0", m.ReconnectCount())
	}

	// Disconnect then connect again => 1 reconnect.
	m.RecordDisconnect(now.Add(1 * time.Minute))
	if m.Connected() {
		t.Fatal("should be disconnected after RecordDisconnect")
	}
	m.RecordConnect(now.Add(2 * time.Minute))
	if m.ReconnectCount() != 1 {
		t.Fatalf("reconnect count = %d, want 1", m.ReconnectCount())
	}

	// Another cycle => 2 reconnects.
	m.RecordDisconnect(now.Add(3 * time.Minute))
	m.RecordConnect(now.Add(4 * time.Minute))
	if m.ReconnectCount() != 2 {
		t.Fatalf("reconnect count = %d, want 2", m.ReconnectCount())
	}
}

func TestGapThreshold(t *testing.T) {
	interval := 5 * time.Minute
	if got := GapThreshold(interval, 1.5); got != 7*time.Minute+30*time.Second {
		t.Errorf("GapThreshold = %v, want 7m30s", got)
	}
	// graceFactor<=0 falls back to default.
	if got := GapThreshold(interval, 0); got != time.Duration(float64(interval)*DefaultWSGraceFactor) {
		t.Errorf("GapThreshold default = %v", got)
	}
	// zero interval => 0.
	if got := GapThreshold(0, 1.5); got != 0 {
		t.Errorf("GapThreshold zero interval = %v, want 0", got)
	}
}

func TestEvaluateKLineGapNoBaseline(t *testing.T) {
	m := NewWSHealthMonitor()
	// No kline recorded yet => never alerts.
	if alert, _ := m.EvaluateKLineGap(time.Now(), 5*time.Minute, 1.5, time.Minute); alert {
		t.Fatal("should not alert without a kline baseline")
	}
}

func TestEvaluateKLineGapWithinThreshold(t *testing.T) {
	m := NewWSHealthMonitor()
	start := time.Now()
	m.RecordKLine(start)

	// 6m gap on a 5m interval with 1.5 factor (threshold 7m30s) => no alert.
	if alert, gap := m.EvaluateKLineGap(start.Add(6*time.Minute), 5*time.Minute, 1.5, time.Minute); alert {
		t.Fatalf("should not alert within threshold, gap=%v", gap)
	}
}

func TestEvaluateKLineGapExceedsThreshold(t *testing.T) {
	m := NewWSHealthMonitor()
	start := time.Now()
	m.RecordKLine(start)

	// 8m gap on a 5m interval => exceeds 7m30s threshold => alert.
	alert, gap := m.EvaluateKLineGap(start.Add(8*time.Minute), 5*time.Minute, 1.5, time.Minute)
	if !alert {
		t.Fatalf("should alert when gap exceeds threshold, gap=%v", gap)
	}
	if gap != 8*time.Minute {
		t.Errorf("gap = %v, want 8m", gap)
	}
}

func TestEvaluateKLineGapThrottle(t *testing.T) {
	m := NewWSHealthMonitor()
	start := time.Now()
	m.RecordKLine(start)

	minAlert := 5 * time.Minute

	// First breach alerts.
	if alert, _ := m.EvaluateKLineGap(start.Add(8*time.Minute), 5*time.Minute, 1.5, minAlert); !alert {
		t.Fatal("first breach should alert")
	}
	// Second breach 2m later is throttled.
	if alert, _ := m.EvaluateKLineGap(start.Add(10*time.Minute), 5*time.Minute, 1.5, minAlert); alert {
		t.Fatal("second breach within throttle window should not alert")
	}
	// After the throttle window elapses, alerts again.
	if alert, _ := m.EvaluateKLineGap(start.Add(14*time.Minute), 5*time.Minute, 1.5, minAlert); !alert {
		t.Fatal("breach after throttle window should alert again")
	}
}

func TestEvaluateKLineGapResetByNewKLine(t *testing.T) {
	m := NewWSHealthMonitor()
	start := time.Now()
	m.RecordKLine(start)

	// A fresh kline resets the baseline so the gap shrinks below threshold.
	m.RecordKLine(start.Add(6 * time.Minute))
	if alert, gap := m.EvaluateKLineGap(start.Add(7*time.Minute), 5*time.Minute, 1.5, time.Minute); alert {
		t.Fatalf("new kline should reset gap, unexpected alert gap=%v", gap)
	}
}

func TestShouldAlertDisconnect(t *testing.T) {
	m := NewWSHealthMonitor()
	now := time.Now()

	// Connected => never alerts.
	m.RecordConnect(now)
	if m.ShouldAlertDisconnect(now, DefaultWSDisconnectAlertBase, DefaultWSDisconnectAlertMax) {
		t.Fatal("should not alert while connected")
	}

	// Disconnect => first check alerts immediately.
	m.RecordDisconnect(now.Add(time.Second))
	if !m.ShouldAlertDisconnect(now.Add(2*time.Second), DefaultWSDisconnectAlertBase, DefaultWSDisconnectAlertMax) {
		t.Fatal("first disconnect check should alert")
	}
	// Immediately after => throttled (backoff not elapsed).
	if m.ShouldAlertDisconnect(now.Add(3*time.Second), DefaultWSDisconnectAlertBase, DefaultWSDisconnectAlertMax) {
		t.Fatal("second immediate check should be throttled")
	}
	// After backoff elapses => alerts again.
	if !m.ShouldAlertDisconnect(now.Add(2*time.Minute), DefaultWSDisconnectAlertBase, DefaultWSDisconnectAlertMax) {
		t.Fatal("check after backoff should alert")
	}

	// Reconnect resets disconnect alert throttling.
	m.RecordConnect(now.Add(3 * time.Minute))
	if m.ShouldAlertDisconnect(now.Add(3*time.Minute), DefaultWSDisconnectAlertBase, DefaultWSDisconnectAlertMax) {
		t.Fatal("should not alert after reconnect")
	}
}
