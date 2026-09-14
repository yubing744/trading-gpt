package exchange

import (
	"testing"

	"github.com/c9s/bbgo/pkg/fixedpoint"
)

func fp(s string) fixedpoint.Value { return fixedpoint.MustNewFromString(s) }

func TestResolveCloseBaseQuantity(t *testing.T) {
	minQty := fp("0.01")

	cases := []struct {
		name         string
		internalBase string
		authNetBase  string
		hasAuth      bool
		percentage   string
		wantQty      string
		wantSkipDust bool
	}{
		{
			// #89 core: internal base is gross/stale and larger than the real
			// exchange net. Sizing off internal would oversell and flip short;
			// clamping to the authoritative net closes exactly what is held.
			name:         "long_internal_gross_exceeds_exchange_net",
			internalBase: "1.6046",
			authNetBase:  "0.8",
			hasAuth:      true,
			percentage:   "1",
			wantQty:      "0.8",
			wantSkipDust: false,
		},
		{
			// Exchange already flat: internal still thinks long -> must not
			// place any sell (that sell is the unintended short).
			name:         "long_internal_but_exchange_flat",
			internalBase: "0.8",
			authNetBase:  "0",
			hasAuth:      true,
			percentage:   "1",
			wantQty:      "0",
			wantSkipDust: true,
		},
		{
			// Exchange net is sub-minimum dust: skip, do not reverse.
			name:         "exchange_net_is_dust",
			internalBase: "0.8",
			authNetBase:  "0.004",
			hasAuth:      true,
			percentage:   "1",
			wantQty:      "0",
			wantSkipDust: true,
		},
		{
			name:         "short_net_clamped_to_magnitude",
			internalBase: "-2.0",
			authNetBase:  "-1.2",
			hasAuth:      true,
			percentage:   "1",
			wantQty:      "1.2",
			wantSkipDust: false,
		},
		{
			name:         "partial_close_half_of_net",
			internalBase: "1.0",
			authNetBase:  "1.0",
			hasAuth:      true,
			percentage:   "0.5",
			wantQty:      "0.5",
			wantSkipDust: false,
		},
		{
			// No authoritative source -> fall back to internal base, still clamped.
			name:         "no_authoritative_falls_back_to_internal",
			internalBase: "0.8",
			authNetBase:  "0",
			hasAuth:      false,
			percentage:   "1",
			wantQty:      "0.8",
			wantSkipDust: false,
		},
		{
			// Defensive: percentage>1 must never exceed the net magnitude.
			name:         "percentage_over_one_clamped",
			internalBase: "1.0",
			authNetBase:  "1.0",
			hasAuth:      true,
			percentage:   "1.5",
			wantQty:      "1.0",
			wantSkipDust: false,
		},
		{
			name:         "nonpositive_percentage_skips",
			internalBase: "1.0",
			authNetBase:  "1.0",
			hasAuth:      true,
			percentage:   "0",
			wantQty:      "0",
			wantSkipDust: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := resolveCloseBaseQuantity(
				fp(tc.internalBase),
				fp(tc.authNetBase),
				tc.hasAuth,
				fp(tc.percentage),
				minQty,
			)
			if got.SkipDust != tc.wantSkipDust {
				t.Fatalf("SkipDust = %v, want %v", got.SkipDust, tc.wantSkipDust)
			}
			if got.BaseQuantity.Compare(fp(tc.wantQty)) != 0 {
				t.Fatalf("BaseQuantity = %v, want %v", got.BaseQuantity, tc.wantQty)
			}
			// Invariant: a resolved close never exceeds the net magnitude.
			if !got.SkipDust {
				netMag := fp(tc.authNetBase).Abs()
				if tc.hasAuth && got.BaseQuantity.Compare(netMag) > 0 {
					t.Fatalf("close qty %v exceeds authoritative net %v (would flip)", got.BaseQuantity, netMag)
				}
			}
		})
	}
}
