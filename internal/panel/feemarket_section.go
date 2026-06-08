package panel

import (
	"fmt"
	"html"
	"strings"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func writeFeemarket(w Writer, d model.Report) {
	w.Section("6. FEE MARKET")
	w.Hint("Base fee adjusts each block from prior-block gas demand vs the network target (EIP-1559-style `x/feemarket`).")
	writeFeemarketSection(w, d)
	w.BlankLine()
}

func writeFeemarketSection(w Writer, d model.Report) {
	ex := buildFeemarketExplain(d)
	writeFeemarketHero(w, ex)
	writeFeemarketFlow(w, ex)
	writeFeemarketReference(w, ex)
}

func writeFeemarketHero(w Writer, ex FeemarketExplain) {
	summary := ex.SummaryLine
	if summary == "" {
		summary = "—"
	}
	heroMeta := ex.HeroLine
	if heroMeta == "" {
		heroMeta = "—"
	}
	w.WriteHTML(fmt.Sprintf(
		`<div class="fee-hero">`+
			`<div class="fee-hero__top">`+
			`<div class="fee-badge fee-badge--%s">%s</div>`+
			`<span class="fee-hero-meta">%s</span>`+
			`</div>`+
			`<p class="fee-hero-summary">%s</p>`+
			`</div>`,
		html.EscapeString(ex.TrafficClass),
		html.EscapeString(ex.TrafficLabel),
		inlineHTML(heroMeta),
		inlineHTML(summary),
	))
}

func writeFeemarketFlow(w Writer, ex FeemarketExplain) {
	if len(ex.FlowSteps) == 0 {
		return
	}
	var b strings.Builder
	b.WriteString(`<div class="fee-flow" role="list">`)
	for i, step := range ex.FlowSteps {
		if i > 0 {
			b.WriteString(`<div class="fee-flow__connector" aria-hidden="true">→</div>`)
		}
		accent := step.Accent
		if accent == "" {
			accent = "default"
		}
		stepCls := fmt.Sprintf("fee-flow__step fee-flow__step--%s", html.EscapeString(accent))
		if step.Headline == "—" && accent == "wallet" {
			stepCls += " fee-flow__step--empty"
		}
		fmt.Fprintf(&b, `<div class="%s" role="listitem">`, stepCls)
		b.WriteString(`<div class="fee-flow__header">`)
		fmt.Fprintf(&b, `<div class="fee-flow__label">%s</div>`, html.EscapeString(step.Label))
		fmt.Fprintf(&b, `<div class="fee-flow__title">%s</div>`, html.EscapeString(step.Title))
		b.WriteString(`</div><div class="fee-flow__body">`)
		if step.Headline != "" {
			fmt.Fprintf(&b, `<div class="fee-flow__headline">%s</div>`, inlineHTML(step.Headline))
		}
		if step.ShowMeter {
			b.WriteString(feemarketDemandMeter(ex))
		}
		if len(step.Values) > 0 {
			b.WriteString(feeFlowValuesHTML(step.Values))
		}
		b.WriteString(`</div></div>`)
	}
	b.WriteString(`</div>`)
	w.WriteHTML(b.String())
}

func feeFlowValuesHTML(values []string) string {
	var b strings.Builder
	b.WriteString(`<dl class="fee-flow__kv">`)
	for _, v := range values {
		key, val, caption := splitFlowValue(v)
		b.WriteString(`<div class="fee-flow__kv-row">`)
		if key != "" {
			fmt.Fprintf(&b, `<dt>%s</dt>`, html.EscapeString(key))
		}
		b.WriteString(`<dd>`)
		if val != "" {
			b.WriteString(inlineHTML(val))
		}
		if caption != "" {
			fmt.Fprintf(&b, `<span class="fee-flow__caption">%s</span>`, inlineHTML(caption))
		}
		b.WriteString(`</dd></div>`)
	}
	b.WriteString(`</dl>`)
	return b.String()
}

// splitFlowValue parses "key: value _(caption)_" or bare "_(caption)_" flow lines.
func splitFlowValue(s string) (key, val, caption string) {
	loc := inlineEmRE.FindStringSubmatchIndex(s)
	if loc != nil {
		caption = s[loc[2]:loc[3]]
		s = strings.TrimSpace(s[:loc[0]])
	}
	if idx := strings.Index(s, ": "); idx >= 0 {
		return strings.TrimSpace(s[:idx]), strings.TrimSpace(s[idx+2:]), caption
	}
	return "", s, caption
}

func feemarketDemandMeter(ex FeemarketExplain) string {
	barPct := ex.LoadBarPct
	if barPct < 0 {
		barPct = 0
	}
	meterLabel := ex.UtilizationPct
	if meterLabel == "" {
		meterLabel = "—"
	}
	return fmt.Sprintf(
		`<div class="fee-meter" role="meter" aria-valuenow="%.0f" aria-valuemin="0" aria-valuemax="100" aria-label="Demand vs capacity">`+
			`<div class="fee-meter__label"><span>Demand vs capacity</span><span>%s</span></div>`+
			`<div class="fee-meter__track"><div class="fee-meter__fill" style="width:%.1f%%"></div></div>`+
			`</div>`,
		barPct, html.EscapeString(meterLabel), barPct,
	)
}

func writeFeemarketReference(w Writer, ex FeemarketExplain) {
	w.Details("feemarket-ref", "Parameters, formulas & data sources", func(w Writer) {
		w.Hint("`gas_used`, W → CometBFT GET /block_results (H−1); W fallback → REST GET /cosmos/evm/feemarket/v1/block_gas; `base_fee` → REST GET …/base_fee; params → REST GET …/params; `eth_gasPrice` → JSON-RPC eth_gasPrice.")
		if len(ex.VariableRows) > 0 {
			w.Subsection("Symbols")
			w.Table([]string{"Symbol", "Meaning", "Live value"}, ex.VariableRows)
		}
		if len(ex.ParamRows) > 0 {
			w.Subsection("Chain parameters")
			w.Table([]string{"Setting", "Value", "Meaning"}, ex.ParamRows)
		}
		if len(ex.FormulaBlocks) > 0 {
			w.Subsection("Formulas")
			writeFeemarketFormulas(w, ex.FormulaBlocks)
		}
	})
}

func writeFeemarketFormulas(w Writer, blocks []string) {
	for _, block := range blocks {
		w.WriteHTML(`<pre class="fee-formula"><code>` + html.EscapeString(block) + `</code></pre>`)
	}
}
