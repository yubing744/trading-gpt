package exchange

import "github.com/c9s/bbgo/pkg/fixedpoint"

// closeSizing is the resolved sizing decision for closing (part of) a position.
type closeSizing struct {
	// BaseQuantity is the absolute base-asset quantity to submit for the close
	// order. It is always <= the authoritative net position magnitude, so a
	// market close can never cross zero and open an unintended opposite position.
	BaseQuantity fixedpoint.Value
	// SkipDust is true when the authoritative net position is below the market
	// minimum quantity. In that case no closing order should be placed, because
	// a reversing order for a dust remainder is exactly the #89 oversell/flip bug.
	SkipDust bool
}

// resolveCloseBaseQuantity decides how much base asset to sell/buy to close a
// position, using the exchange-authoritative net position when it is available
// instead of the (potentially stale or gross) internally tracked base.
//
// Background (#89): handleCleanPosition queries the exchange for the real net
// position to evaluate TP/SL triggers, but ClosePosition then sized the close
// order off the internal ent.position base. When the internal base was larger
// than the exchange net (gross vs net divergence, or leftover dust from a prior
// round), the market close sold more than was actually held and flipped the
// account into an unintended opposite position (observed: a stop-loss close
// oversold and stayed short ~22h).
//
// The fix keeps direction handling in the caller and only governs magnitude:
//   - prefer the authoritative net magnitude when hasAuthoritative is true;
//   - never size larger than that magnitude (clamp defends against percentage>1);
//   - treat a sub-minimum remainder as dust and skip the order entirely.
func resolveCloseBaseQuantity(
	internalBase fixedpoint.Value,
	authoritativeNetBase fixedpoint.Value,
	hasAuthoritative bool,
	percentage fixedpoint.Value,
	minQuantity fixedpoint.Value,
) closeSizing {
	effective := internalBase
	if hasAuthoritative {
		effective = authoritativeNetBase
	}

	magnitude := effective.Abs()

	// Nothing (or only dust) actually held: do not place a reversing order.
	if magnitude.Sign() == 0 || magnitude.Compare(minQuantity) < 0 {
		return closeSizing{BaseQuantity: fixedpoint.Zero, SkipDust: true}
	}

	pct := percentage
	if pct.Sign() <= 0 {
		return closeSizing{BaseQuantity: fixedpoint.Zero, SkipDust: true}
	}
	if pct.Compare(fixedpoint.One) > 0 {
		pct = fixedpoint.One
	}

	qty := magnitude.Mul(pct)
	// Defensive clamp: a close order must never exceed the net position.
	if qty.Compare(magnitude) > 0 {
		qty = magnitude
	}

	return closeSizing{BaseQuantity: qty, SkipDust: false}
}
