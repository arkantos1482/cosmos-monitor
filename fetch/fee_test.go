package fetch

import "testing"

func TestFormatFeeAmount(t *testing.T) {
	tests := []struct {
		raw, denom, want string
	}{
		{"0", "apmt", "0"},
		{"0.000000000000000000", "apmt", "0"},
		{"1000000000", "apmt", "1.00e-09 PMT"},
		{"0", "apmt", "0"},
		{"0.5 PMT", "apmt", "0.5 PMT"},
	}
	for _, tc := range tests {
		if got := FormatFeeAmount(tc.raw, tc.denom); got != tc.want {
			t.Errorf("FormatFeeAmount(%q,%q) = %q, want %q", tc.raw, tc.denom, got, tc.want)
		}
	}
}
