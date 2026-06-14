package panel

import (
	"fmt"
	"html"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func writeDistribution(w Writer, d model.Report) {
	w.Section("5. DISTRIBUTION")
	writeEmbeddedSectionIntro(w, "Unclaimed staking rewards, where those coins sit on-chain, and x/distribution params from BeginBlock fee routing.")
	writeDistributionSummary(w, d, SummaryEmbedded)

	w.Hint("`unclaimed total`, `delegator share`, `operator commission` → derived (Σ per-validator outstanding_rewards + commission from distribution REST; delegator share adjusted when bank escrow matches Σ outstanding); `distribution escrow` → REST bank balance for the distribution module account; `community pool` → REST GET /cosmos/distribution/v1beta1/community_pool; `community_tax`, `withdraw_addr_enabled` → REST GET /cosmos/distribution/v1beta1/params; `escrow check` → derived (bank distribution balance vs unclaimed total); validator table → REST outstanding_rewards + commission per valoper; `comm. rate`, local validator → REST GET /cosmos/staking/v1beta1/validators; local identity → CometBFT GET /status.")

	if d.Local.IsValidator {
		w.Subsection("This validator")
		writeDistributionLocal(w, d.Local)
	}

	w.Subsection("Network-wide")
	w.WriteHTML(distributionDomainCardsHTML(d))
	writeDistributionValidatorTable(w, d)

	writeSectionSources(w, ViewDistribution, d)
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
	if total := distributionUnclaimedTotal(d); total != "" {
		writeDistributionSummaryKPI(w, "unclaimed total", total, "")
	}
	if d.UnclaimedDelegator != "" {
		writeDistributionSummaryKPI(w, "delegator share", d.UnclaimedDelegator, "")
	}
	if d.UnclaimedCommission != "" {
		writeDistributionSummaryKPI(w, "operator commission", d.UnclaimedCommission, "")
	}
	if bal := distributionModuleBalance(d); bal != "" {
		writeDistributionSummaryKPI(w, "distribution escrow", bal, "")
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
	if html := localUnclaimedBreakdownHTML(lv); html != "" {
		w.RowHTML("unclaimed rewards", html, "")
	} else {
		w.Row("unclaimed rewards", "—")
	}
	if lv.Commission > 0 {
		w.Row("commission rate", fmt.Sprintf("%.1f%% of rewards before delegator split", lv.Commission))
	}
}
