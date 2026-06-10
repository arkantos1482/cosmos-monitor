package feemarket

import (
	"strconv"

	"github.com/arkantos1482/cosmos-monitor/internal/fetch"
)

// TransferCost formats 21,000 × base fee for display.
func TransferCost(baseFeeRaw, denom string) string {
	if baseFeeRaw == "" || baseFeeRaw == "0" {
		return "—"
	}
	v, err := strconv.ParseFloat(baseFeeRaw, 64)
	if err != nil || v <= 0 {
		return "—"
	}
	total := v * float64(StandardTransferGas)
	return fetch.FormatFeeAmount(strconv.FormatFloat(total, 'f', 0, 64), denom)
}
