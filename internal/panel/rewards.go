package panel

import (
	"fmt"
	"html"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func writeRewards(w Writer, d model.Report) {
	w.Section("3. REWARDS")
	writeRewardsSummary(w, d, SummaryEmbedded)
	w.Em("Block reward sources (PMT emissions and mint inflation) and this validator's unclaimed share. Distribution routing → § Economics.")

	w.Layer("Chain")
	w.WriteHTML(economicsDomainCardsHTML(d, false))
	writeEconomicsLedger(w, d)
	w.Hint(rewardsSourcesHint())

	if d.Local.IsValidator {
		w.Layer("This validator")
		writeRewardsLocalValidator(w, d)
	}
	w.BlankLine()
}

func writeRewardsSummary(w Writer, d model.Report, mode SummaryMode) {
	summaryWrapStart(w, mode, "rewards")
	if mode == SummaryOverviewClickable {
		writeRewardsCompactSummary(w, d)
	} else {
		w.WriteHTML(`<div class="eco-summary">`)
		w.WriteHTML(economicsDomainCardsHTML(d, true))
		w.WriteHTML(`</div>`)
	}
	summaryWrapEnd(w, mode)
}

func writeRewardsCompactSummary(w Writer, d model.Report) {
	writeRewardsChainCompactSummary(w, d)
	if d.Local.IsValidator && d.Local.Outstanding != "" {
		w.WriteHTML(fmt.Sprintf(
			`<div class="eco-summary__row">Outstanding: %s</div>`,
			html.EscapeString(d.Local.Outstanding)))
	}
}

func writeRewardsChainCompactSummary(w Writer, d model.Report) {
	w.WriteHTML(`<div class="eco-summary eco-summary--compact">`)
	writeRewardsChainStatusRows(w, d)
	w.WriteHTML(`</div>`)
}

func writeRewardsChainStatusRows(w Writer, d model.Report) {
	pmtStatus := "disabled"
	if d.PMTEnabled {
		switch {
		case d.PMTPoolEmpty:
			pmtStatus = "pool empty"
		case d.PMTRate != "":
			pmtStatus = d.PMTRate
		default:
			pmtStatus = "enabled"
		}
	}
	w.WriteHTML(fmt.Sprintf(`<div class="eco-summary__row">PMT: %s</div>`, html.EscapeString(pmtStatus)))

	inflStatus := "off"
	if d.Inflation > 0 {
		inflStatus = fmt.Sprintf("%.2f%%", d.Inflation)
	}
	w.WriteHTML(fmt.Sprintf(`<div class="eco-summary__row">Inflation: %s</div>`, html.EscapeString(inflStatus)))
}

func writeRewardsLocalValidator(w Writer, d model.Report) {
	lv := d.Local
	w.Subsection("Unclaimed")
	w.Hint("`outstanding rewards`, `commission earned` → REST GET /cosmos/distribution/v1beta1/validators/{valoper}/outstanding_rewards, …/commission; `per-block` → derived (network reward flow × VP% × commission).")
	if lv.Outstanding != "" {
		w.Row("outstanding rewards", lv.Outstanding+"  _(total unclaimed — x/distribution)_")
	} else {
		w.Row("outstanding rewards", "–")
	}
	if lv.CommissionEarned != "" {
		w.Row("commission earned", lv.CommissionEarned+"  _(unclaimed validator commission)_")
	} else {
		w.Row("commission earned", "–")
	}
	if op, del, _, ok := localValidatorPerBlockRewards(d); ok {
		w.Row("per-block commission", op+fmt.Sprintf("  (%.2f%% VP · %.2f%% commission)", lv.VPPercent, lv.Commission))
		w.Row("per-block delegators", del)
	}
}

func rewardsSourcesHint() string {
	return "`PMT rewards` → REST GET /cosmos/evm/pmtrewards/v1/params; " +
		"`inflation`, `annual provisions` → REST GET /cosmos/mint/v1beta1/inflation, /cosmos/mint/v1beta1/annual-provisions; " +
		"`blocks / year`, mint params → REST GET /cosmos/mint/v1beta1/params; " +
		"`ledger per-block amounts` → derived (PMT rate, mint inflation/block, parent-block fees); " +
		"`fee_collector cleared` → derived (x/bank fee_collector balance)."
}
