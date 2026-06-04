package panel

import (
	"fmt"
	"html"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func writeFeemarketSection(w Writer, d model.Report) {
	ex := buildFeemarketExplain(d)
	writeFeemarketHero(w, ex)
	writeFeemarketKeyMetrics(w, ex)
	if len(ex.ParamRows) > 0 {
		w.Subsection("Chain parameters")
		w.Table([]string{"Setting", "Value", "Note"}, ex.ParamRows)
	}
	w.Subsection("Receipt")
	w.Pre(ex.Receipt)
	writeFeemarketWalletChain(w, ex)
	if len(ex.AdjustmentBullets) > 0 {
		for _, b := range ex.AdjustmentBullets {
			w.ListItem(b)
		}
	}
}

func writeFeemarketHero(w Writer, ex FeemarketExplain) {
	barPct := ex.LoadBarPct
	if barPct < 0 {
		barPct = 0
	}
	meterLabel := ex.UtilizationPct
	if meterLabel == "" {
		meterLabel = "—"
	}
	w.WriteHTML(fmt.Sprintf(
		`<div class="fee-traffic">`+
			`<div class="fee-badge fee-badge--%s">%s</div>`+
			`<div class="fee-meter" role="meter" aria-valuenow="%.0f" aria-valuemin="0" aria-valuemax="100" aria-label="Previous block load vs target">`+
			`<div class="fee-meter__label"><span>Block load vs target</span><span>%s</span></div>`+
			`<div class="fee-meter__track"><div class="fee-meter__fill" style="width:%.1f%%"></div></div>`+
			`</div>`+
			`<p class="fee-hero-line">%s</p>`+
			`</div>`,
		html.EscapeString(ex.TrafficClass),
		html.EscapeString(ex.TrafficLabel),
		barPct,
		html.EscapeString(meterLabel),
		barPct,
		inlineHTML(ex.HeroLine),
	))
}

func writeFeemarketKeyMetrics(w Writer, ex FeemarketExplain) {
	w.WriteHTML(fmt.Sprintf(
		`<dl class="stat-grid fee-key-metrics">`+
			`<div class="stat"><dt>base fee</dt><dd>%s</dd></div>`+
			`<div class="stat"><dt>eth_gasPrice</dt><dd>%s</dd></div>`+
			`<div class="stat"><dt>next adjustment</dt><dd>%s</dd></div>`+
			`</dl>`,
		html.EscapeString(ex.StatBaseFee),
		html.EscapeString(ex.StatGasPrice),
		html.EscapeString(ex.StatNextAdj),
	))
}

func writeFeemarketWalletChain(w Writer, ex FeemarketExplain) {
	w.WriteHTML(
		`<div class="fee-cards">` +
			`<div class="fee-card">` +
			`<h4 class="fee-card__title">What wallets see</h4>` +
			fmt.Sprintf(`<p class="note">%s</p>`, inlineHTML(ex.WalletLine)) +
			`</div>` +
			`<div class="fee-card">` +
			`<h4 class="fee-card__title">What the chain enforces</h4>` +
			fmt.Sprintf(`<p class="note">%s</p>`, inlineHTML(ex.ChainLine)) +
			`</div></div>`,
	)
}
