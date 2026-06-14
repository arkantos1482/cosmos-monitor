package panel

import (
	"fmt"
	"html"
	"strconv"

	"github.com/arkantos1482/cosmos-monitor/internal/feemarket"
)

// fmGasLimitHTML renders consensus max_gas: (-1) ∞ with raw MaxUint64 in small gray text.
func fmGasLimitHTML(s feemarket.State) string {
	if s.UnlimitedBlockGas {
		return `(-1) ∞ <span class="fm-sentinel__raw">` +
			html.EscapeString(formatUint(s.GasLimit)+` gas`) + `</span>`
	}
	if s.GasLimit > 0 {
		return html.EscapeString(formatUint(s.GasLimit) + " gas")
	}
	return "—"
}

// fmGasTargetHTML renders gas target; unlimited max_gas shows "max ÷ N" with raw sentinel underneath.
func fmGasTargetHTML(s feemarket.State) string {
	if s.GasTarget == 0 {
		return "—"
	}
	if s.UnlimitedBlockGas {
		label := "max ÷ elasticity"
		if s.Elasticity > 0 {
			label = fmt.Sprintf("max ÷ %d", s.Elasticity)
		}
		return html.EscapeString(label) + ` <span class="fm-sentinel__raw">` +
			html.EscapeString(formatUint(s.GasTarget)+` gas`) + `</span>`
	}
	return html.EscapeString(formatUint(s.GasTarget) + " gas")
}

func fmDemandVsTargetHTML(s feemarket.State) string {
	if s.UnlimitedBlockGas || s.GasTarget == 0 || s.GasWanted == 0 {
		return ""
	}
	pct := strconv.Itoa(s.UtilPct)
	bar := fmt.Sprintf(`<div class="kpi-bar"><div class="kpi-bar__fill" style="width:%s%%"></div></div>`, pct)
	return html.EscapeString(pct+"%") + " " + bar
}
