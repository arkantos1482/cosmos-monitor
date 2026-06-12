package panel

import (
	"fmt"
	"html"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func writeDistribution(w Writer, d model.Report) {
	w.Section("5. DISTRIBUTION")
	writeEmbeddedSectionIntro(w, "x/distribution params, community pool, module accounts, and per-validator outstanding rewards from BeginBlock fee routing.")
	writeDistributionSummary(w, d, SummaryEmbedded)

	if d.Local.IsValidator {
		w.Subsection("This validator")
		writeDistributionLocal(w, d.Local)
	}

	w.Subsection("Network-wide")
	w.WriteHTML(distributionDomainCardsHTML(d))
	writeDistributionValidatorTable(w, d)

	w.Hint(distributionSourcesHint())
	w.BlankLine()
}

func writeDistributionSummary(w Writer, d model.Report, mode SummaryMode) {
	summaryWrapStart(w, mode, "distribution")
	writeDistributionSummaryBody(w, d)
	summaryWrapEnd(w, mode)
}

func writeDistributionSummaryBody(w Writer, d model.Report) {
	w.WriteHTML(`<div class="dist-summary">`)
	w.WriteHTML(`<div class="dist-summary__kpis">`)
	if d.CommunityPool != "" {
		writeDistributionSummaryKPI(w, "community pool", d.CommunityPool, "")
	}
	if d.CommunityTax != "" {
		tone := ""
		if d.CommunityTaxZero {
			tone = "warn"
		}
		writeDistributionSummaryKPI(w, "community tax", d.CommunityTax, tone)
	}
	if total := distributionUnclaimedTotal(d); total != "" {
		writeDistributionSummaryKPI(w, "unclaimed", total, "")
	}
	if status := distributionFeeCollectorStatus(d); status != "—" {
		writeDistributionSummaryKPI(w, "fee_collector", status, distributionFeeCollectorTone(d))
	}
	w.WriteHTML(`</div></div>`)
}

func writeDistributionSummaryKPI(w Writer, label, value, tone string) {
	if value == "" {
		return
	}
	valCls := "dist-summary__kpi-val"
	if tone != "" {
		valCls += " dist-summary__kpi-val--" + tone
	}
	w.WriteHTML(fmt.Sprintf(
		`<div class="dist-summary__kpi"><span class="dist-summary__kpi-label">%s</span>`+
			`<span class="%s">%s</span></div>`,
		html.EscapeString(label), valCls, html.EscapeString(value)))
}

func writeDistributionLocal(w Writer, lv model.LocalValidator) {
	w.Hint("`outstanding rewards`, `commission` → REST GET /cosmos/distribution/v1beta1/validators/{valoper}/outstanding_rewards, …/commission.")
	if lv.Outstanding != "" {
		w.Row("outstanding rewards", lv.Outstanding+"  _(unclaimed delegator share)_")
	} else {
		w.Row("outstanding rewards", "—")
	}
	if lv.CommissionEarned != "" {
		w.Row("commission", lv.CommissionEarned+"  _(unclaimed validator commission)_")
	} else {
		w.Row("commission", "—")
	}
}

func distributionSourcesHint() string {
	return "`community_tax`, `withdraw_addr_enabled` → REST GET /cosmos/distribution/v1beta1/params; " +
		"`community pool` → REST GET /cosmos/distribution/v1beta1/community_pool; " +
		"`outstanding rewards`, `commission` → REST GET /cosmos/distribution/v1beta1/validators/{valoper}/outstanding_rewards, …/commission; " +
		"`module account balances` → REST GET /cosmos/bank/v1beta1/balances/{address}; " +
		"`module account addresses` → REST GET /cosmos/auth/v1beta1/module_accounts."
}
