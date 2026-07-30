package fetch

import (
	"strings"
	"testing"

	"cosmossdk.io/math"
)

func TestFormatAmount(t *testing.T) {
	tests := []struct {
		v    float64
		want string
	}{
		{0, "0"},
		{1.5e12, "1.5T"},
		{4e8, "400M"},
		{1500, "1.5K"},
		{42.5, "42.5"},
		{0.001, "0.001"},
		{1e-8, "0.00000001"},
		{1e-9, "0.000000001"},
		{7e-18, "7e-18"},
		{-3.2e6, "-3.2M"},
		// Near-round compact values must not snap to a cleaner neighbor.
		{399999600, "399.9996M"},
		{99999900, "99.9999M"},
		{1234567, "1.234567M"},
		{1.23e6, "1.23M"},
	}
	for _, tc := range tests {
		if got := FormatAmount(tc.v); got != tc.want {
			t.Errorf("FormatAmount(%v) = %q, want %q", tc.v, got, tc.want)
		}
	}
}

func TestFormatPct(t *testing.T) {
	tests := []struct {
		v    float64
		want string
	}{
		{0, "0%"},
		{55.5, "55.5%"},
		{0.01, "0.01%"},
		{0.0099, "0.0099%"},
		{0.00002, "0.00002%"},
		{-1.25, "-1.25%"},
	}
	for _, tc := range tests {
		if got := FormatPct(tc.v); got != tc.want {
			t.Errorf("FormatPct(%v) = %q, want %q", tc.v, got, tc.want)
		}
	}
}

func TestSumRawAmounts(t *testing.T) {
	got := SumRawAmounts("99999900000000000000000000", "100000000000000000000")
	want := "100000000000000000000000000"
	if got != want {
		t.Fatalf("SumRawAmounts = %q, want %q", got, want)
	}
}

func TestFormatShares(t *testing.T) {
	got := FormatShares("100000000000000000000", "apmt")
	if got != "100" {
		t.Errorf("FormatShares = %q, want 100", got)
	}
}

func TestFormatCoinLargeInteger(t *testing.T) {
	// 400e24 apmt → 400M PMT (must not lose precision via float64 parse of raw string).
	got := FormatCoin("400000000000000000000000000", "apmt")
	want := "400M PMT"
	if got != want {
		t.Errorf("FormatCoin(large apmt) = %q, want %q", got, want)
	}
}

func TestFormatCoinNearRoundMillion(t *testing.T) {
	got := FormatCoin("399999600000000000000000000", "apmt")
	want := "399.9996M PMT"
	if got != want {
		t.Errorf("FormatCoin(near-round) = %q, want %q", got, want)
	}
}

func TestFormatFeeStepTruncates(t *testing.T) {
	raw := math.LegacyNewDec(7).QuoInt(math.NewIntFromUint64(8))
	got := FormatFeeStep(raw, "apmt")
	if got == "0" || got == "0 apmt" {
		t.Fatalf("expected truncate hint, got %q", got)
	}
	if !strings.Contains(got, "truncates to 0") {
		t.Fatalf("FormatFeeStep = %q", got)
	}
}

func TestFormatFeeDecInteger(t *testing.T) {
	got := FormatFeeDec(math.LegacyNewDec(7), "apmt")
	if got != "7 apmt" {
		t.Fatalf("FormatFeeDec = %q", got)
	}
}

func TestFormatFeeAmount(t *testing.T) {
	tests := []struct {
		raw, denom, want string
	}{
		{"0", "apmt", "0"},
		{"0.000000000000000000", "apmt", "0"},
		{"1000000000", "apmt", "1000000000 apmt"},
		{"0.000000000000000007", "apmt", "7e-18 PMT"},
		{"0.5 PMT", "apmt", "0.5 PMT"},
	}
	for _, tc := range tests {
		if got := FormatFeeAmount(tc.raw, tc.denom); got != tc.want {
			t.Errorf("FormatFeeAmount(%q,%q) = %q, want %q", tc.raw, tc.denom, got, tc.want)
		}
	}
}
