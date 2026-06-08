package panel

import (
	"fmt"
	"html"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func writeFeemarketSection(w Writer, d model.Report) {
	ex := buildFeemarketExplain(d)
	w.Hint("`gas_used`, W → CometBFT `block_results` (H−1); W fallback → `GET /cosmos/evm/feemarket/v1/block_gas`; `base_fee` → `…/base_fee`; params → `…/params`; `eth_gasPrice` → EVM JSON-RPC.")
	writeFeemarketHero(w, ex, d)
	if len(ex.VariableRows) > 0 {
		w.Subsection("Variables")
		w.Hint("`W` → fee-market hero sources; `target` → `x/feemarket` params + `block_gas`; `base_fee` → `GET /cosmos/evm/feemarket/v1/base_fee`; `utilization` → derived from `W` / `target`.")
		w.Table([]string{"Symbol", "Meaning", "Live value"}, ex.VariableRows)
	}
	if len(ex.FormulaBlocks) > 0 {
		w.Subsection("Formulas")
		w.Hint("`base fee adjustment` → `x/feemarket` EIP-1559-style formula (constants from params below).")
		writeFeemarketFormulas(w, ex.FormulaBlocks)
	}
	if len(ex.ParamRows) > 0 {
		w.Subsection("Params")
		w.Hint("`elasticity`, `base_fee_change_denominator`, `min_gas_price`, … → `GET /cosmos/evm/feemarket/v1/params`.")
		w.Table([]string{"Setting", "Value", "Meaning"}, ex.ParamRows)
	}
}

func writeFeemarketHero(w Writer, ex FeemarketExplain, d model.Report) {
	var meterHTML string
	if !ex.HideLoadMeter {
		barPct := ex.LoadBarPct
		if barPct < 0 {
			barPct = 0
		}
		meterLabel := ex.UtilizationPct
		if meterLabel == "" {
			meterLabel = "—"
		}
		meterHTML = fmt.Sprintf(
			`<div class="fee-meter" role="meter" aria-valuenow="%.0f" aria-valuemin="0" aria-valuemax="100" aria-label="Previous block load vs target">`+
				`<div class="fee-meter__label"><span>W / target</span><span>%s</span></div>`+
				`<div class="fee-meter__track"><div class="fee-meter__fill" style="width:%.1f%%"></div></div>`+
				`</div>`,
			barPct, html.EscapeString(meterLabel), barPct,
		)
	}

	util := ex.UtilizationPct
	if util == "" {
		util = "—"
	}
	nextAdj := ex.NextAdj
	if nextAdj == "" {
		nextAdj = "—"
	}
	baseFee := d.BaseFee
	if baseFee == "" {
		baseFee = "—"
	}

	kpiHTML := fmt.Sprintf(
		`<div class="fee-key-metrics">`+
			`<div class="fee-kpi"><div class="fee-kpi__label">Utilization</div><div class="fee-kpi__value">%s</div></div>`+
			`<div class="fee-kpi"><div class="fee-kpi__label">Next adjustment</div><div class="fee-kpi__value">%s</div></div>`+
			`<div class="fee-kpi"><div class="fee-kpi__label">Base fee</div><div class="fee-kpi__value">%s</div></div>`+
			`</div>`,
		html.EscapeString(util),
		html.EscapeString(nextAdj),
		html.EscapeString(baseFee),
	)

	w.WriteHTML(fmt.Sprintf(
		`<div class="fee-hero">`+
			`<div class="fee-traffic">`+
			`<div class="fee-badge fee-badge--%s">%s</div>`+
			`%s`+
			`<p class="fee-hero-line">%s</p>`+
			`%s`+
			`</div></div>`,
		html.EscapeString(ex.TrafficClass),
		html.EscapeString(ex.TrafficLabel),
		meterHTML,
		inlineHTML(ex.HeroLine),
		kpiHTML,
	))
}

func writeFeemarketFormulas(w Writer, blocks []string) {
	for _, block := range blocks {
		w.WriteHTML(`<pre class="fee-formula"><code>` + html.EscapeString(block) + `</code></pre>`)
	}
}
