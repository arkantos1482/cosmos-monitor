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

	w.Subsection("Unclaimed rewards")
	writeDistributionNetworkUnclaimed(w, d)
	writeDistributionValidatorTable(w, d)

	w.Subsection("Treasury & params")
	w.WriteHTML(distributionParamsCardHTML(d))

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
	if d.Local.IsValidator {
		w.WriteHTML(`<div class="dist-summary__columns">`)
		writeDistributionSummaryScope(w, "This validator", unclaimedStackFromLocal(d.Local))
		writeDistributionSummaryScope(w, "Network", unclaimedStackFromNetwork(d))
		w.WriteHTML(`</div>`)
	} else {
		writeDistributionSummaryScope(w, "Network", unclaimedStackFromNetwork(d))
	}
	w.WriteHTML(`</div>`)
}

func writeDistributionSummaryScope(w Writer, title string, stack unclaimedStack) {
	if stack.empty() {
		return
	}
	w.WriteHTML(`<div class="dist-summary__scope">`)
	w.WriteHTML(fmt.Sprintf(`<div class="dist-summary__scope-label">%s</div>`, html.EscapeString(title)))
	w.WriteHTML(unclaimedStackHTML(stack))
	w.WriteHTML(`</div>`)
}

func writeDistributionLocal(w Writer, lv model.LocalValidator) {
	if markup := localUnclaimedBreakdownHTML(lv); markup != "" {
		w.WriteHTML(`<div class="dist-section-unclaimed">` + markup + `</div>`)
	} else {
		w.Row("unclaimed rewards", "—")
	}
	if lv.Commission > 0 {
		w.Row("commission rate", fmt.Sprintf("%.1f%% of rewards before delegator split", lv.Commission))
	}
}

func writeDistributionNetworkUnclaimed(w Writer, d model.Report) {
	if markup := networkUnclaimedBreakdownHTML(d); markup != "" {
		w.WriteHTML(`<div class="dist-section-unclaimed">` + markup + `</div>`)
	} else {
		w.Row("unclaimed rewards", "—")
	}
	if html := distributionEscrowBlockHTML(d); html != "" {
		w.WriteHTML(html)
	}
}
