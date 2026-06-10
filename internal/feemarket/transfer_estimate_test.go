package feemarket

import "testing"

func TestTransferCost(t *testing.T) {
	tests := []struct {
		baseFee, denom, want string
	}{
		{"", "apmt", "—"},
		{"0", "apmt", "—"},
		{"1000", "apmt", "21000000 apmt"},
		{"0.000000000000000007", "apmt", "1.47e-13 PMT"},
	}
	for _, tc := range tests {
		if got := TransferCost(tc.baseFee, tc.denom); got != tc.want {
			t.Errorf("TransferCost(%q, %q) = %q, want %q", tc.baseFee, tc.denom, got, tc.want)
		}
	}
}
