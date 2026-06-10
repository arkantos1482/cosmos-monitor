package feemarket

import (
	"github.com/arkantos1482/cosmos-monitor/internal/fetch"
)

// TransferCost formats 21,000 × base fee for display.
func TransferCost(baseFeeRaw, denom string) string {
	if baseFeeRaw == "" || baseFeeRaw == "0" {
		return "—"
	}
	d, ok := ParseLegacyDec(baseFeeRaw)
	if !ok || !d.IsPositive() {
		return "—"
	}
	total := d.MulInt64(StandardTransferGas)
	return fetch.FormatFeeDec(total, denom)
}
