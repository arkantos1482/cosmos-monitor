package panel

import (
	"fmt"
	"html"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func writeFeemarketSection(w Writer, d model.Report) {
	ex := buildFeemarketExplain(d)
	w.Hint("`gas_used`, stored W → CometBFT `block_results` (height−1); W fallback → `GET /cosmos/evm/feemarket/v1/block_gas`; `base_fee` → `…/base_fee`; chain params → `…/params`; `eth_gasPrice` → EVM JSON-RPC. Payout path is in the Overview diagram above.")
	writeFeemarketHero(w, ex)
	if len(ex.LastBlockRows) > 0 {
		w.Subsection("Last block (N−1)")
		w.Table([]string{"Field", "Value", "Source"}, ex.LastBlockRows)
	}
	if ex.FormulaLine != "" {
		w.Subsection("Formula")
		w.Em(ex.FormulaLine)
	}
	if len(ex.ThisBlockRows) > 0 {
		w.Subsection("This block (N)")
		w.Table([]string{"Field", "Value", "Source"}, ex.ThisBlockRows)
	}
	if len(ex.ParamRows) > 0 {
		w.Subsection("Params")
		w.Table([]string{"Setting", "Value", "Meaning"}, ex.ParamRows)
	}
	writeFeemarketWalletChain(w, ex)
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
			`<div class="fee-meter__label"><span>W / target</span><span>%s</span></div>`+
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
