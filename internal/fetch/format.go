package fetch

import (
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"

	sdkmath "cosmossdk.io/math"
)

// FormatAmount renders a human-scale number (already denom-adjusted) for dashboards.
// Large values use compact SI-style suffixes (K, M, B, T); small values use trimmed
// fixed-point or scientific notation when needed.
func FormatAmount(v float64) string {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return "—"
	}
	if v == 0 {
		return "0"
	}
	sign := ""
	if v < 0 {
		sign = "-"
		v = -v
	}
	return sign + formatAmountAbs(v)
}

// FormatAmountUnit is FormatAmount with an optional unit suffix (e.g. "PMT", "PMT/year").
func FormatAmountUnit(v float64, unit string) string {
	s := FormatAmount(v)
	if unit == "" {
		return s
	}
	return s + " " + unit
}

func formatAmountAbs(v float64) string {
	for _, s := range []struct {
		div float64
		suf string
	}{
		{1e12, "T"},
		{1e9, "B"},
		{1e6, "M"},
		{1e3, "K"},
	} {
		if v >= s.div {
			return fmt.Sprintf("%.2f%s", v/s.div, s.suf)
		}
	}
	if v >= 1 {
		return trimTrailingZeros(fmt.Sprintf("%.4f", v))
	}
	if v >= 1e-4 {
		return trimTrailingZeros(fmt.Sprintf("%.6f", v))
	}
	if v > 0 && v < 1e-4 {
		dec := int(math.Ceil(-math.Log10(v))) + 2
		if dec > 21 {
			dec = 21
		}
		s := trimTrailingZeros(fmt.Sprintf("%.*f", dec, v))
		if len(s) > 12 {
			return formatSci(v)
		}
		return s
	}
	return formatSci(v)
}

func formatSci(v float64) string {
	return strconv.FormatFloat(v, 'g', -1, 64)
}

func trimTrailingZeros(s string) string {
	if !strings.Contains(s, ".") {
		return s
	}
	s = strings.TrimRight(s, "0")
	return strings.TrimRight(s, ".")
}

// coinToDisplay converts a raw on-chain amount + denom to display float and ticker.
// Integer strings use math/big to avoid float64 precision loss on large Cosmos amounts.
func coinToDisplay(rawAmount, denom string) (float64, string) {
	rawAmount = strings.TrimSpace(rawAmount)
	if rawAmount == "" {
		return 0, displayDenom(denom)
	}
	if strings.Contains(rawAmount, ".") {
		v, _ := strconv.ParseFloat(rawAmount, 64)
		return convertDenom(v, denom)
	}
	amt := new(big.Int)
	if _, ok := amt.SetString(rawAmount, 10); !ok {
		v, _ := strconv.ParseFloat(rawAmount, 64)
		return convertDenom(v, denom)
	}
	div, disp := denomDivisor(denom)
	if div.Cmp(big.NewInt(1)) == 0 {
		f, _ := new(big.Rat).SetInt(amt).Float64()
		return f, disp
	}
	r := new(big.Rat).SetFrac(amt, div)
	f, _ := r.Float64()
	return f, disp
}

func displayDenom(denom string) string {
	if denom == "" {
		return ""
	}
	switch denom[0] {
	case 'a', 'n', 'u', 'm':
		return strings.ToUpper(denom[1:])
	}
	return strings.ToUpper(denom)
}

func denomDivisor(denom string) (*big.Int, string) {
	if len(denom) == 0 {
		return big.NewInt(1), ""
	}
	switch denom[0] {
	case 'a':
		return bigPow10(18), displayDenom(denom)
	case 'n':
		return bigPow10(9), displayDenom(denom)
	case 'u':
		return bigPow10(6), displayDenom(denom)
	case 'm':
		return bigPow10(3), displayDenom(denom)
	}
	return big.NewInt(1), displayDenom(denom)
}

func bigPow10(n int) *big.Int {
	return new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(n)), nil)
}

// FormatFeeDec formats a Cosmos legacy decimal fee amount using FormatFeeAmount.
// Whole-number values use the integer code path so apmt amounts match REST raw strings (e.g. "7" → "7 apmt").
func FormatFeeDec(d sdkmath.LegacyDec, denom string) string {
	if d.IsNil() {
		return "—"
	}
	if d.TruncateDec().Equal(d) {
		return FormatFeeAmount(d.TruncateInt().String(), denom)
	}
	s := strings.TrimRight(strings.TrimRight(d.String(), "0"), ".")
	return FormatFeeAmount(s, denom)
}

// FormatFeeStep formats a fee delta and marks values that truncate to zero at integer apmt precision.
func FormatFeeStep(d sdkmath.LegacyDec, denom string) string {
	if d.IsNil() || d.IsZero() {
		return FormatFeeAmount("0", denom)
	}
	if d.TruncateDec().IsZero() && d.IsPositive() {
		return fmt.Sprintf("<%s (truncates to 0)", FormatFeeDec(d, denom))
	}
	return FormatFeeDec(d, denom)
}
