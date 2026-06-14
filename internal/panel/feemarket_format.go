package panel

import (
	"fmt"
	"html"
	"strconv"
	"strings"

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

func fmMechanicsVarsHTML(s feemarket.State) string {
	var rows []struct{ name, value, hint string }
	if s.GasUsed > 0 || s.TxGasWanted > 0 || s.GasWanted > 0 {
		rows = append(rows, struct{ name, value, hint string }{
			"gas_used", fmGasAmount(s.GasUsed), "parent block execution (Σ tx gas_used)",
		})
		rows = append(rows, struct{ name, value, hint string }{
			"gas_wanted", fmGasAmount(s.TxGasWanted), "parent block mempool limits (Σ tx gas_wanted)",
		})
		if s.MinGasMult != "" {
			rows = append(rows, struct{ name, value, hint string }{
				"min_gas_multiplier", html.EscapeString(s.MinGasMult), "scales gas_wanted before max()",
			})
		}
		if s.GasWanted > 0 {
			rows = append(rows, struct{ name, value, hint string }{
				"W", fmGasAmount(s.GasWanted), "max(gas_used, gas_wanted × min_gas_multiplier) — formula input",
			})
		}
	}
	if s.GasTarget > 0 {
		rows = append(rows, struct{ name, value, hint string }{
			"target", fmGasTargetHTML(s), "block gas limit ÷ elasticity_multiplier",
		})
	}
	if s.BaseFee != "" {
		rows = append(rows, struct{ name, value, hint string }{
			"base_fee", html.EscapeString(s.BaseFee), "parent block base fee",
		})
	}
	if s.ChangeDenom > 0 {
		rows = append(rows, struct{ name, value, hint string }{
			"denom", strconv.FormatInt(s.ChangeDenom, 10), "base_fee_change_denominator",
		})
	}
	if s.MinGasPrice != "" {
		rows = append(rows, struct{ name, value, hint string }{
			"min_gas_price", html.EscapeString(s.MinGasPrice), "floor when base fee decreases",
		})
	}
	if len(rows) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString(`<dl class="fm-mechanics__vars">`)
	for _, r := range rows {
		b.WriteString(`<div class="fm-mechanics__var">`)
		b.WriteString(`<dt>` + html.EscapeString(r.name) + `</dt>`)
		b.WriteString(`<dd><span class="fm-mechanics__var-val">` + r.value + `</span>`)
		if r.hint != "" {
			b.WriteString(` <span class="fm-mechanics__var-hint">` + html.EscapeString(r.hint) + `</span>`)
		}
		b.WriteString(`</dd></div>`)
	}
	b.WriteString(`</dl>`)
	return b.String()
}

func fmGasAmount(n uint64) string {
	if n == 0 {
		return "0 gas"
	}
	return html.EscapeString(formatUint(n) + " gas")
}
