package panel

import (
	"fmt"
	"html"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func writeRewards(w Writer, d model.Report) {
	w.Section("3. REWARDS")
	writeRewardsSummary(w, d, SummaryEmbedded)
	w.Em("Block reward sources: PMT emissions and mint inflation, with per-block estimates for this validator.")

	w.Layer("Network-wide")
	w.WriteHTML(economicsDomainCardsHTML(d, false))
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
	w.Subsection("Per-block estimates")
	w.Hint("`per-block` → derived (network reward flow × VP% × commission).")
	if op, del, _, ok := localValidatorPerBlockRewards(d); ok {
		w.Row("per-block commission", op+fmt.Sprintf("  (%.2f%% VP · %.2f%% commission)", lv.VPPercent, lv.Commission))
		w.Row("per-block delegators", del)
	}
}

func rewardsSourcesHint() string {
	return "`PMT rewards` → REST GET /cosmos/evm/pmtrewards/v1/params; " +
		"`inflation`, `annual provisions` → REST GET /cosmos/mint/v1beta1/inflation, /cosmos/mint/v1beta1/annual-provisions; " +
		"`blocks / year`, mint params → REST GET /cosmos/mint/v1beta1/params; " +
		"`per-block amounts` → derived (PMT rate, mint inflation/block, parent-block fees)."
}
