package exchange

import (
	"testing"
	"time"

	"github.com/c9s/bbgo/pkg/fixedpoint"
	"github.com/c9s/bbgo/pkg/types"
)

func TestOpenIntent(t *testing.T) {
	if got := OpenIntent(types.SideTypeBuy); got != IntentOpenLong {
		t.Fatalf("OpenIntent(Buy) = %q, want %q", got, IntentOpenLong)
	}
	if got := OpenIntent(types.SideTypeSell); got != IntentOpenShort {
		t.Fatalf("OpenIntent(Sell) = %q, want %q", got, IntentOpenShort)
	}
}

func TestCloseIntentDiscriminatesByPercentage(t *testing.T) {
	full := CloseIntent(fixedpoint.One)
	half := CloseIntent(fixedpoint.NewFromFloat(0.5))

	if full == half {
		t.Fatalf("expected distinct intents for distinct percentages, both = %q", full)
	}

	// Same percentage must be stable (deterministic).
	if again := CloseIntent(fixedpoint.One); again != full {
		t.Fatalf("CloseIntent(1) not deterministic: %q vs %q", again, full)
	}
}

func TestBuildIdempotencyKey(t *testing.T) {
	cycle := time.Unix(1_700_000_000, 0)

	k1 := BuildIdempotencyKey("BTCUSDT", IntentOpenLong, cycle)
	if k1 == "" {
		t.Fatal("expected non-empty key for valid inputs")
	}

	// Deterministic for identical inputs.
	if k2 := BuildIdempotencyKey("BTCUSDT", IntentOpenLong, cycle); k1 != k2 {
		t.Fatalf("key not deterministic: %q vs %q", k1, k2)
	}

	// Changing any field changes the key.
	if k := BuildIdempotencyKey("ETHUSDT", IntentOpenLong, cycle); k == k1 {
		t.Fatal("expected different key for different symbol")
	}
	if k := BuildIdempotencyKey("BTCUSDT", IntentOpenShort, cycle); k == k1 {
		t.Fatal("expected different key for different intent")
	}
	if k := BuildIdempotencyKey("BTCUSDT", IntentOpenLong, cycle.Add(time.Minute)); k == k1 {
		t.Fatal("expected different key for different cycle")
	}

	// Zero cycle or empty symbol => empty key (no idempotency context).
	if k := BuildIdempotencyKey("BTCUSDT", IntentOpenLong, time.Time{}); k != "" {
		t.Fatalf("expected empty key for zero cycle, got %q", k)
	}
	if k := BuildIdempotencyKey("", IntentOpenLong, cycle); k != "" {
		t.Fatalf("expected empty key for empty symbol, got %q", k)
	}
}

// Same intent submitted twice in the same cycle: only the first proceeds.
func TestGuardSameIntentSubmittedOnce(t *testing.T) {
	g := newOrderDedupGuard()
	cycle := time.Unix(1_700_000_000, 0)

	if _, first := g.claim("BTCUSDT", IntentOpenLong, cycle); !first {
		t.Fatal("first claim should proceed")
	}
	if _, first := g.claim("BTCUSDT", IntentOpenLong, cycle); first {
		t.Fatal("duplicate claim in same cycle should be suppressed")
	}
	// A third identical retry must also stay suppressed.
	if _, first := g.claim("BTCUSDT", IntentOpenLong, cycle); first {
		t.Fatal("further duplicate claims in same cycle should stay suppressed")
	}
}

// Different intents in the same cycle each proceed independently.
func TestGuardDifferentIntentsEachSubmit(t *testing.T) {
	g := newOrderDedupGuard()
	cycle := time.Unix(1_700_000_000, 0)

	if _, first := g.claim("BTCUSDT", IntentOpenLong, cycle); !first {
		t.Fatal("open_long should proceed")
	}
	if _, first := g.claim("BTCUSDT", IntentOpenShort, cycle); !first {
		t.Fatal("open_short is a distinct intent and should proceed")
	}
	if _, first := g.claim("BTCUSDT", CloseIntent(fixedpoint.One), cycle); !first {
		t.Fatal("close is a distinct intent and should proceed")
	}
	if _, first := g.claim("BTCUSDT", CloseIntent(fixedpoint.NewFromFloat(0.5)), cycle); !first {
		t.Fatal("a different close percentage is a distinct intent and should proceed")
	}
	// A different symbol is also independent.
	if _, first := g.claim("ETHUSDT", IntentOpenLong, cycle); !first {
		t.Fatal("different symbol should proceed")
	}
}

// The same intent is allowed again once the kline cycle advances.
func TestGuardAllowsResubmitAcrossCycles(t *testing.T) {
	g := newOrderDedupGuard()
	c1 := time.Unix(1_700_000_000, 0)
	c2 := c1.Add(5 * time.Minute)

	if _, first := g.claim("BTCUSDT", IntentOpenLong, c1); !first {
		t.Fatal("first claim in cycle 1 should proceed")
	}
	if _, first := g.claim("BTCUSDT", IntentOpenLong, c1); first {
		t.Fatal("duplicate in cycle 1 should be suppressed")
	}
	if _, first := g.claim("BTCUSDT", IntentOpenLong, c2); !first {
		t.Fatal("same intent in a new cycle should proceed again")
	}
	if _, first := g.claim("BTCUSDT", IntentOpenLong, c2); first {
		t.Fatal("duplicate in cycle 2 should be suppressed")
	}
}

// With no idempotency context (empty key), every call proceeds and nothing is
// recorded, preserving legacy behaviour for callers without kline data.
func TestGuardEmptyKeyAlwaysProceeds(t *testing.T) {
	g := newOrderDedupGuard()

	if _, first := g.claim("BTCUSDT", IntentOpenLong, time.Time{}); !first {
		t.Fatal("zero cycle should proceed (no dedup context)")
	}
	if _, first := g.claim("BTCUSDT", IntentOpenLong, time.Time{}); !first {
		t.Fatal("zero cycle should keep proceeding (never recorded)")
	}
	if _, first := g.claim("", IntentOpenLong, time.Unix(1_700_000_000, 0)); !first {
		t.Fatal("empty symbol should proceed (no dedup context)")
	}
}

func TestTryClaimEmptyKey(t *testing.T) {
	g := newOrderDedupGuard()
	if !g.tryClaim("", time.Unix(1, 0)) {
		t.Fatal("empty key must always proceed")
	}
	if !g.tryClaim("", time.Unix(2, 0)) {
		t.Fatal("empty key must always proceed regardless of cycle")
	}
}
