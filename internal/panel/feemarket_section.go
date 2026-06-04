package panel

import (
	"fmt"
	"html"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func writeFeemarketSection(w Writer, d model.Report) {
	ex := buildFeemarketExplain(d)

	w.StrongLine(ex.SummaryLine)
	writeFeemarketTraffic(w, ex)
	writeFeemarketStatusDashboard(w, ex)
	w.Subsection("Receipt walkthrough")
	w.Pre(ex.Receipt)
	writeFeemarketWalletChainCards(w, ex)
}

func writeFeemarketTraffic(w Writer, ex FeemarketExplain) {
	barPct := ex.LoadBarPct
	if barPct < 0 {
		barPct = 0
	}
	loadLabel := ex.StatLastLoad
	if ex.UtilizationPct != "" {
		loadLabel = ex.UtilizationPct + " of target"
	}
	w.WriteHTML(fmt.Sprintf(
		`<div class="fee-traffic">`+
			`<div class="fee-badge fee-badge--%s">%s</div>`+
			`<div class="fee-meter" role="meter" aria-valuenow="%.0f" aria-valuemin="0" aria-valuemax="100" aria-label="Previous block load vs target">`+
			`<div class="fee-meter__label"><span>Block load vs target</span><span>%s</span></div>`+
			`<div class="fee-meter__track"><div class="fee-meter__fill" style="width:%.1f%%"></div></div>`+
			`</div>`+
			`<dl class="stat-grid fee-traffic-stats">`+
			`<div class="stat"><dt>base fee now</dt><dd>%s</dd></div>`+
			`<div class="stat"><dt>eth_gasPrice</dt><dd>%s</dd></div>`+
			`<div class="stat"><dt>next adjustment</dt><dd>%s</dd></div>`+
			`</dl></div>`,
		html.EscapeString(ex.TrafficClass),
		html.EscapeString(ex.TrafficLabel),
		barPct,
		html.EscapeString(loadLabel),
		barPct,
		html.EscapeString(ex.StatBaseFee),
		html.EscapeString(ex.StatGasPrice),
		html.EscapeString(ex.StatNextAdj),
	))
}

func writeFeemarketStatusDashboard(w Writer, ex FeemarketExplain) {
	w.Hint(ex.Hint)
	w.Row("base fee now", ex.StatBaseFee)
	w.Row("eth_gasPrice", ex.StatGasPrice)
	w.Row("next adjustment", ex.StatNextAdj+"  _("+ex.Verdict+")_")
	w.Row("last block load", ex.StatLastLoad)
	if len(ex.ParamRows) > 0 {
		w.Subsection("Chain parameters")
		w.Table([]string{"Setting", "Value", "What it does"}, ex.ParamRows)
	}
}

func writeFeemarketWalletChainCards(w Writer, ex FeemarketExplain) {
	w.WriteHTML(
		`<div class="fee-cards">` +
			`<div class="fee-card">` +
			`<h4 class="fee-card__title">What wallets see</h4>` +
			`<dl class="stat-grid">` +
			fmt.Sprintf(`<div class="stat"><dt>eth_gasPrice</dt><dd>%s</dd></div>`, html.EscapeString(ex.WalletGasPrice)) +
			`</dl>` +
			fmt.Sprintf(`<p class="note">%s</p>`, inlineHTML(ex.WalletPayNote)) +
			`</div>` +
			`<div class="fee-card">` +
			`<h4 class="fee-card__title">What the chain enforces</h4>` +
			`<dl class="stat-grid">` +
			fmt.Sprintf(`<div class="stat"><dt>base_fee</dt><dd>%s</dd></div>`, html.EscapeString(ex.ChainBaseFee)) +
			fmt.Sprintf(`<div class="stat"><dt>no_base_fee</dt><dd>%s</dd></div>`, html.EscapeString(ex.ChainNoBaseFee)) +
			`</dl>` +
			fmt.Sprintf(`<p class="note">%s</p>`, inlineHTML(ex.ChainDemandNote)) +
			`</div></div>`,
	)
	w.Subsection("Adjustment logic")
	for _, b := range ex.AdjustmentBullets {
		w.ListItem(b)
	}
}
