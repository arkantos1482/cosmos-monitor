package panel

import (
	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func writeDistribution(w Writer, d model.Report) {
	w.Section("5. DISTRIBUTION")
	writeEmbeddedSectionIntro(w, "Per-block reward ledger, `fee_collector` and `x/distribution` routing, community tax and pool, and unclaimed balances network-wide and for this validator.")
	writeDistributionSummary(w, d, SummaryEmbedded)

	w.Layer("Network-wide")
	writeDistributionLedger(w, d)
	writeDistributionDetails(w, d)
	w.Hint(distributionSourcesHint())

	if d.Local.IsValidator {
		w.Layer("This validator")
		writeDistributionLocalValidator(w, d)
	}
	w.BlankLine()
}

func writeDistributionSummary(w Writer, d model.Report, mode SummaryMode) {
	summaryWrapStart(w, mode, "distribution")
	w.WriteHTML(`<div class="dist-summary">`)
	writeDistributionSummaryRows(w, d)
	w.WriteHTML(`</div>`)
	summaryWrapEnd(w, mode)
}

func writeDistributionLocalValidator(w Writer, d model.Report) {
	lv := d.Local
	w.Subsection("Unclaimed")
	w.Hint("`outstanding rewards`, `commission earned` → REST GET /cosmos/distribution/v1beta1/validators/{valoper}/outstanding_rewards, …/commission.")
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
}

func distributionSourcesHint() string {
	return "`community tax`, `community pool` → REST GET /cosmos/distribution/v1beta1/params, /cosmos/distribution/v1beta1/community_pool; " +
		"`unclaimed delegator`, `unclaimed commission` → REST GET /cosmos/distribution/v1beta1/validators/{valoper}/outstanding_rewards, …/commission (summed across validators); " +
		"`module account balances` → REST GET /cosmos/bank/v1beta1/balances/{address}; " +
		"`module account addresses` → REST GET /cosmos/auth/v1beta1/module_accounts; " +
		"`fee_collector cleared`, `unclaimed check` → derived (x/bank balances, outstanding sums); " +
		"`ledger per-block amounts` → derived (PMT rate, mint inflation/block, parent-block fees)."
}
