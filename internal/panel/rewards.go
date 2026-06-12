package panel

import (
	"fmt"
	"html"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func writeRewards(w Writer, d model.Report) {
	w.Section("2. REWARDS")
	writeRewardsSummary(w, d, SummaryEmbedded)
	w.Em("Block reward sources (PMT emissions, mint inflation), distribution routing through `fee_collector` and `x/distribution`, and unclaimed balances chain-wide and for this validator.")

	w.Layer("Network-wide")
	w.WriteHTML(economicsDomainCardsHTML(d, false))
	writeRewardsLedger(w, d)
	writeRewardsDistribution(w, d)
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
	writeRewardsDistributionSummaryRows(w, d)
	w.WriteHTML(`</div>`)
	if d.Local.IsValidator && d.Local.Outstanding != "" {
		w.WriteHTML(fmt.Sprintf(
			`<div class="eco-summary__row">Outstanding: %s</div>`,
			html.EscapeString(d.Local.Outstanding)))
	}
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
		"`community tax`, `community pool` → REST GET /cosmos/distribution/v1beta1/params, /cosmos/distribution/v1beta1/community_pool; " +
		"`unclaimed delegator`, `unclaimed commission` → REST GET /cosmos/distribution/v1beta1/validators/{valoper}/outstanding_rewards, …/commission (summed across validators); " +
		"`module account balances` → REST GET /cosmos/bank/v1beta1/balances/{address}; " +
		"`module account addresses` → REST GET /cosmos/auth/v1beta1/module_accounts; " +
		"`fee_collector cleared`, `unclaimed check` → derived (x/bank balances, outstanding sums)."
}
