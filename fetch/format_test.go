package fetch

import "testing"

func TestFormatAmount(t *testing.T) {
	tests := []struct {
		v    float64
		want string
	}{
		{0, "0"},
		{1.5e12, "1.50T"},
		{4e8, "400.00M"},
		{1500, "1.50K"},
		{42.5, "42.5"},
		{0.001, "0.001"},
		{1e-8, "1.00e-08"},
		{1e-9, "1.00e-09"},
		{-3.2e6, "-3.20M"},
	}
	for _, tc := range tests {
		if got := FormatAmount(tc.v); got != tc.want {
			t.Errorf("FormatAmount(%v) = %q, want %q", tc.v, got, tc.want)
		}
	}
}

func TestFormatCoinLargeInteger(t *testing.T) {
	// 400e24 apmt → 400M PMT (must not lose precision via float64 parse of raw string).
	got := FormatCoin("400000000000000000000000000", "apmt")
	want := "400.00M PMT"
	if got != want {
		t.Errorf("FormatCoin(large apmt) = %q, want %q", got, want)
	}
}

func TestFormatFeeAmount(t *testing.T) {
	tests := []struct {
		raw, denom, want string
	}{
		{"0", "apmt", "0"},
		{"0.000000000000000000", "apmt", "0"},
		{"1000000000", "apmt", "1.00e-09 PMT"},
		{"0.5 PMT", "apmt", "0.5 PMT"},
	}
	for _, tc := range tests {
		if got := FormatFeeAmount(tc.raw, tc.denom); got != tc.want {
			t.Errorf("FormatFeeAmount(%q,%q) = %q, want %q", tc.raw, tc.denom, got, tc.want)
		}
	}
}
