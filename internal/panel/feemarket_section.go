package panel

import (
	"fmt"
	"html"
	"strings"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func writeFeemarket(w Writer, d model.Report) {
	w.Section("6. FEE MARKET")
	w.Em("Base fee adjusts each block from prior-block gas demand vs the network target (EIP-1559-style `x/feemarket`).")
	writeFeemarketSection(w, d)
	w.BlankLine()
}

func writeFeemarketSection(w Writer, d model.Report) {
	ex := buildFeemarketExplain(d)
	writeFeemarketHero(w, ex)
	writeFeemarketPipeline(w, ex)
	writeFeemarketCards(w, ex)
	writeFeemarketReference(w, ex)
}

func writeFeemarketHero(w Writer, ex FeemarketExplain) {
	summary := ex.SummaryLine
	if summary == "" {
		summary = "—"
	}
	w.WriteHTML(fmt.Sprintf(
		`<div class="fee-hero">`+
			`<div class="fee-traffic">`+
			`<div class="fee-badge fee-badge--%s">%s</div>`+
			`<p class="fee-hero-summary">%s</p>`+
			`<p class="fee-hero-line">%s</p>`+
			`</div></div>`,
		html.EscapeString(ex.TrafficClass),
		html.EscapeString(ex.TrafficLabel),
		inlineHTML(summary),
		inlineHTML(ex.HeroLine),
	))
}

func writeFeemarketPipeline(w Writer, ex FeemarketExplain) {
	if len(ex.PipelineSteps) == 0 {
		return
	}
	var b strings.Builder
	b.WriteString(`<div class="fee-pipeline" role="list">`)
	for i, step := range ex.PipelineSteps {
		if i > 0 {
			b.WriteString(`<div class="fee-pipeline__connector" aria-hidden="true">→</div>`)
		}
		b.WriteString(`<div class="fee-pipeline__step" role="listitem">`)
		fmt.Fprintf(&b, `<div class="fee-pipeline__label">%s</div>`, html.EscapeString(step.Label))
		fmt.Fprintf(&b, `<div class="fee-pipeline__title">%s</div>`, html.EscapeString(step.Title))
		b.WriteString(`<div class="fee-pipeline__values">`)
		for _, v := range step.Values {
			fmt.Fprintf(&b, `<div class="fee-pipeline__value">%s</div>`, inlineHTML(v))
		}
		b.WriteString(`</div></div>`)
	}
	b.WriteString(`</div>`)
	w.WriteHTML(b.String())
}

func writeFeemarketCards(w Writer, ex FeemarketExplain) {
	if len(ex.Cards) == 0 {
		return
	}
	var b strings.Builder
	b.WriteString(`<div class="fee-cards">`)
	for _, card := range ex.Cards {
		accent := card.Accent
		if accent == "" {
			accent = "default"
		}
		fmt.Fprintf(&b, `<div class="fee-card fee-card--%s">`, html.EscapeString(accent))
		fmt.Fprintf(&b, `<h4 class="fee-card__title">%s</h4>`, html.EscapeString(card.Title))
		fmt.Fprintf(&b, `<div class="fee-card__primary">%s</div>`, inlineHTML(card.Primary))
		if card.ShowMeter {
			b.WriteString(feemarketDemandMeter(ex))
		}
		if card.FormulaBlock != "" {
			b.WriteString(`<pre class="fee-formula fee-formula--inline"><code>`)
			b.WriteString(html.EscapeString(card.FormulaBlock))
			b.WriteString(`</code></pre>`)
		}
		if len(card.Lines) > 0 {
			b.WriteString(`<ul class="fee-card__lines">`)
			for _, line := range card.Lines {
				fmt.Fprintf(&b, `<li>%s</li>`, inlineHTML(line))
			}
			b.WriteString(`</ul>`)
		}
		if card.Caption != "" {
			fmt.Fprintf(&b, `<p class="fee-card__caption">%s</p>`, inlineHTML(card.Caption))
		}
		b.WriteString(`</div>`)
	}
	b.WriteString(`</div>`)
	w.WriteHTML(b.String())
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
