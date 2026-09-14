package exchange

import (
	"fmt"
	"sync"
	"time"

	"github.com/c9s/bbgo/pkg/fixedpoint"
	"github.com/c9s/bbgo/pkg/types"
)

// OrderIntent identifies the semantic operation an order submission represents.
// Two submissions that share the same intent within the same kline cycle are
// treated as duplicates (typically caused by a network retry or a repeated
// signal dispatch) and only the first one is allowed through.
type OrderIntent string

const (
	IntentOpenLong  OrderIntent = "open_long"
	IntentOpenShort OrderIntent = "open_short"
	IntentClose     OrderIntent = "close"
)

// OpenIntent maps an order side to the corresponding open-position intent.
func OpenIntent(side types.SideType) OrderIntent {
	if side == types.SideTypeSell {
		return IntentOpenShort
	}
	return IntentOpenLong
}

// CloseIntent builds a close intent discriminated by the close percentage so
// that genuinely distinct partial closes within one cycle are not wrongly
// blocked, while identical close retries are deduplicated.
func CloseIntent(percentage fixedpoint.Value) OrderIntent {
	return OrderIntent(fmt.Sprintf("%s:%s", IntentClose, percentage.String()))
}

// BuildIdempotencyKey returns a deterministic key for an order submission.
//
// The key is stable across retries of the same intent within the same kline
// cycle and changes when the symbol, the intent, or the cycle changes. cycle is
// normalised to the kline start time in unix milliseconds. A zero cycle or an
// empty symbol yields an empty key, which signals "no idempotency context is
// available" and disables deduplication for that call (preserving legacy
// behaviour for callers without kline data).
func BuildIdempotencyKey(symbol string, intent OrderIntent, cycle time.Time) string {
	if cycle.IsZero() || symbol == "" {
		return ""
	}

	return fmt.Sprintf("%s|%s|%d", symbol, intent, cycle.UnixMilli())
}

// orderDedupGuard tracks which idempotency keys have already been claimed in the
// current kline cycle. It is safe for concurrent use.
//
// The guard claims a key *before* the order is submitted and never releases it
// within the cycle, even if the submission returns an error. This is the
// conservative choice for #94: the most dangerous failure mode is a submission
// that times out locally but actually reached the exchange; releasing the claim
// on error would let the retry create a real duplicate order. By keeping the
// claim, a duplicate dispatch is suppressed for the rest of the cycle and the
// strategy simply re-evaluates on the next kline.
type orderDedupGuard struct {
	mu    sync.Mutex
	cycle time.Time
	seen  map[string]struct{}
}

func newOrderDedupGuard() *orderDedupGuard {
	return &orderDedupGuard{seen: make(map[string]struct{})}
}

// tryClaim records key for the given cycle and reports whether the caller should
// proceed with the submission. It returns true the first time a key is seen
// within a cycle and false for subsequent duplicates. When the cycle advances,
// previously seen keys are evicted so the same intent may run again. An empty key
// (no idempotency context) always returns true and is never recorded.
func (g *orderDedupGuard) tryClaim(key string, cycle time.Time) bool {
	if key == "" {
		return true
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	if !g.cycle.Equal(cycle) {
		g.cycle = cycle
		g.seen = make(map[string]struct{})
	}

	if _, ok := g.seen[key]; ok {
		return false
	}

	g.seen[key] = struct{}{}
	return true
}

// claim builds the idempotency key for the given intent and cycle and attempts
// to claim it. It returns the key (for logging) and whether this is the first
// submission for that key within the cycle.
func (g *orderDedupGuard) claim(symbol string, intent OrderIntent, cycle time.Time) (string, bool) {
	key := BuildIdempotencyKey(symbol, intent, cycle)
	return key, g.tryClaim(key, cycle)
}
