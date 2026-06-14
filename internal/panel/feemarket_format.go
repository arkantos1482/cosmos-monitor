package panel

import (
	"fmt"
	"html"
	"strconv"
	"strings"

	"github.com/arkantos1482/cosmos-monitor/internal/feemarket"
	"github.com/arkantos1482/cosmos-monitor/internal/fetch"
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
	if !s.EIP1559On && s.BaseFee == "" && s.GasTarget == 0 {
		return ""
	}
	var rows []fmMechVar
	// W = max(gas_used, gas_wanted × min_gas_multiplier)
	rows = append(rows, fmMechVar{"gas_used", fmGasAmount(s.GasUsed), "parent block execution (Σ tx gas_used)"})
	rows = append(rows, fmMechVar{"gas_wanted", fmGasAmount(s.TxGasWanted), "parent block mempool limits (Σ tx gas_wanted)"})
	rows = append(rows, fmMechVar{"min_gas_multiplier", fmMechScalar(s.MinGasMult), "scales gas_wanted before max()"})
	rows = append(rows, fmMechVar{"W", fmGasAmount(s.GasWanted), "max(gas_used, gas_wanted × min_gas_multiplier)"})
	// target = block_gas_limit ÷ elasticity_multiplier
	rows = append(rows, fmMechVar{"block_gas_limit", fmGasLimitHTML(s), "consensus block max_gas"})
	rows = append(rows, fmMechVar{"elasticity_multiplier", fmMechInt(s.Elasticity), "target = block_gas_limit ÷ this"})
	rows = append(rows, fmMechVar{"target", fmGasTargetOrDash(s), "gas target for EIP-1559 adjustment"})
	// base_fee ± base_fee × |W−target| / target / denom
	rows = append(rows, fmMechVar{"|W−target|", fmMechanicsDelta(s), "absolute demand gap used in the formula"})
	rows = append(rows, fmMechVar{"base_fee", fmMechText(s.BaseFee), "parent block base fee"})
	rows = append(rows, fmMechVar{"denom", fmMechInt(s.ChangeDenom), "base_fee_change_denominator"})
	if s.Denom != "" {
		rows = append(rows, fmMechVar{"min_fee_step", html.EscapeString(fetch.FormatFeeAmount("1", s.Denom)), "minimum base-fee increase per block (1 apmt)"})
	}
	rows = append(rows, fmMechVar{"min_gas_price", fmMechText(s.MinGasPrice), "floor when base fee decreases"})
	if s.HasProjection && s.ProjectedRaw != "" {
		rows = append(rows, fmMechVar{"projected_base_fee", html.EscapeString(fetch.FormatFeeAmount(s.ProjectedRaw, s.Denom)), "computed next-block base fee"})
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

type fmMechVar struct {
	name, value, hint string
}

func fmMechanicsDelta(s feemarket.State) string {
	if s.GasWanted == 0 || s.GasTarget == 0 {
		return "—"
	}
	var d uint64
	if s.GasWanted > s.GasTarget {
		d = s.GasWanted - s.GasTarget
	} else {
		d = s.GasTarget - s.GasWanted
	}
	return fmGasAmount(d)
}

func fmGasTargetOrDash(s feemarket.State) string {
	if s.GasTarget == 0 {
		return "—"
	}
	return fmGasTargetHTML(s)
}

func fmMechScalar(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return "—"
	}
	return html.EscapeString(v)
}

func fmMechText(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return "—"
	}
	return html.EscapeString(v)
}

func fmMechInt(v int64) string {
	if v <= 0 {
		return "—"
	}
	return html.EscapeString(strconv.FormatInt(v, 10))
}

func fmGasAmount(n uint64) string {
	if n == 0 {
		return "0 gas"
	}
	return html.EscapeString(formatUint(n) + " gas")
}
