package panel

import (
	"fmt"
	"html"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func writeRewards(w Writer, d model.Report) {
	w.Section("4. REWARDS")
	writeRewardsSummary(w, d, SummaryEmbedded)
	w.Em("Block reward sources: PMT emissions and mint inflation, with per-block estimates for this validator.")

	if d.Local.IsValidator {
		w.Subsection("This validator")
		writeRewardsLocalValidator(w, d)
	}

	w.WriteHTML(ecoDomainsWrap(
		pmtRewardsCardHTML(d, false),
		inflationCardHTML(d, false),
	))

	w.Hint(rewardsSourcesHint())
	w.BlankLine()
}

func writeRewardsSummary(w Writer, d model.Report, mode SummaryMode) {
	summaryWrapStart(w, mode, "rewards")
	writeRewardsCompactSummary(w, d)
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
