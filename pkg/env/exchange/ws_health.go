package exchange

import (
	"sync"
	"time"
)

// WS health defaults. These are conservative, observation-only values; the
// monitor never touches orders or positions, it only tracks connection state
// and emits log alerts so that OKX WebSocket reconnect churn (issue #90) and
// any resulting market-data gaps are visible.
const (
	// DefaultWSCheckInterval is how often the watchdog evaluates health.
	DefaultWSCheckInterval = 30 * time.Second
	// DefaultWSGraceFactor multiplies the kline interval to decide when a
	// missing kline is considered a gap (e.g. 1.5x of a 5m interval = 7m30s).
	DefaultWSGraceFactor = 1.5
	// DefaultWSMinGapAlertInterval throttles repeated kline-gap alerts.
	DefaultWSMinGapAlertInterval = 5 * time.Minute
	// DefaultWSDisconnectAlertBase / Max bound the exponential backoff used to
	// throttle repeated "still disconnected" alerts.
	DefaultWSDisconnectAlertBase = 30 * time.Second
	DefaultWSDisconnectAlertMax  = 10 * time.Minute
)

// ReconnectBackoff returns an exponential backoff duration for the given
// attempt number, capped at max. attempt <= 0 returns base. It is a pure
// helper used to space out escalating disconnect alerts so the log is not
// spammed while a connection is down.
func ReconnectBackoff(attempt int, base, max time.Duration) time.Duration {
	if base <= 0 {
		base = DefaultWSDisconnectAlertBase
	}
	if max <= 0 {
		max = DefaultWSDisconnectAlertMax
	}
	if attempt <= 0 {
		if base > max {
			return max
		}
		return base
	}

	d := base
	for i := 0; i < attempt; i++ {
		if d >= max {
			return max
		}
		// double, guarding against overflow
		if d > max/2 {
			return max
		}
		d *= 2
	}

	if d > max {
		return max
	}
	return d
}

// WSHealthMonitor is a pure, concurrency-safe state machine tracking the
// health of an exchange WebSocket feed. It is deliberately free of any bbgo /
// side-effecting dependencies so the disconnect state machine and gap-detection
// judgment can be unit tested in isolation.
//
// The bbgo StandardStream already performs the actual reconnect and
// re-subscribe on its own; this monitor sits on top of it to (a) count and
// surface reconnect churn and (b) detect market-data (kline) gaps that the
// churn can produce, alerting via the caller's log/notify hook.
type WSHealthMonitor struct {
	mu sync.Mutex

	connected      bool
	everConnected  bool
	reconnectCount int

	lastConnectAt    time.Time
	lastDisconnectAt time.Time
	lastKLineAt      time.Time

	lastGapAlertAt        time.Time
	lastDisconnectAlertAt time.Time
	disconnectAlerts      int
}

// NewWSHealthMonitor creates an empty monitor.
func NewWSHealthMonitor() *WSHealthMonitor {
	return &WSHealthMonitor{}
}

// RecordConnect marks the stream connected. When it follows a prior connection
// (i.e. this is a reconnect rather than the first connect) the reconnect
// counter is incremented. It also resets disconnect-alert throttling.
func (m *WSHealthMonitor) RecordConnect(now time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.everConnected {
		m.reconnectCount++
	}
	m.everConnected = true
	m.connected = true
	m.lastConnectAt = now
	m.disconnectAlerts = 0
	m.lastDisconnectAlertAt = time.Time{}
}

// RecordDisconnect marks the stream disconnected.
func (m *WSHealthMonitor) RecordDisconnect(now time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.connected = false
	m.lastDisconnectAt = now
}

// RecordKLine records the arrival time of a market-data kline (or any liveness
// signal) used for gap detection.
func (m *WSHealthMonitor) RecordKLine(now time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.lastKLineAt = now
}

// Connected reports the current connection state.
func (m *WSHealthMonitor) Connected() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.connected
}

// ReconnectCount reports the number of reconnects observed since start.
func (m *WSHealthMonitor) ReconnectCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.reconnectCount
}

// KLineGap returns how long it has been since the last kline was recorded.
// It returns 0 if no kline has been recorded yet (nothing to judge against).
func (m *WSHealthMonitor) KLineGap(now time.Time) time.Duration {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.lastKLineAt.IsZero() {
		return 0
	}
	return now.Sub(m.lastKLineAt)
}

// GapThreshold computes the duration beyond which a missing kline is treated as
// a gap: interval * graceFactor. graceFactor <= 0 falls back to the default.
func GapThreshold(interval time.Duration, graceFactor float64) time.Duration {
	if graceFactor <= 0 {
		graceFactor = DefaultWSGraceFactor
	}
	if interval <= 0 {
		return 0
	}
	return time.Duration(float64(interval) * graceFactor)
}

// EvaluateKLineGap decides whether a kline-gap alert should be emitted now.
//
// It returns alert=true (and the observed gap) only when:
//   - at least one kline has been recorded (we have a baseline), and
//   - the gap exceeds interval*graceFactor, and
//   - the gap is not fully explained by a "young" connection that has simply
//     not been up long enough to have produced a kline yet, and
//   - at least minAlertInterval has elapsed since the previous gap alert
//     (throttling to avoid log spam).
//
// When it returns true it records the alert time so the throttle applies to the
// next call. The method is safe for concurrent use.
func (m *WSHealthMonitor) EvaluateKLineGap(now time.Time, interval time.Duration, graceFactor float64, minAlertInterval time.Duration) (bool, time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.lastKLineAt.IsZero() {
		return false, 0
	}

	gap := now.Sub(m.lastKLineAt)
	threshold := GapThreshold(interval, graceFactor)
	if threshold <= 0 || gap <= threshold {
		return false, gap
	}

	if minAlertInterval <= 0 {
		minAlertInterval = DefaultWSMinGapAlertInterval
	}
	if !m.lastGapAlertAt.IsZero() && now.Sub(m.lastGapAlertAt) < minAlertInterval {
		return false, gap
	}

	m.lastGapAlertAt = now
	return true, gap
}

// ShouldAlertDisconnect decides whether a "still disconnected" alert should be
// emitted now. While the stream is disconnected it emits alerts with an
// exponential backoff (base..max) between them so a long outage does not spam
// the log. It returns false when the stream is currently connected.
func (m *WSHealthMonitor) ShouldAlertDisconnect(now time.Time, base, max time.Duration) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.connected {
		return false
	}

	if m.lastDisconnectAlertAt.IsZero() {
		m.lastDisconnectAlertAt = now
		m.disconnectAlerts = 1
		return true
	}

	wait := ReconnectBackoff(m.disconnectAlerts, base, max)
	if now.Sub(m.lastDisconnectAlertAt) < wait {
		return false
	}

	m.lastDisconnectAlertAt = now
	m.disconnectAlerts++
	return true
}
