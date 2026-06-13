package panel

import (
	"fmt"
	"html"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
	"github.com/arkantos1482/cosmos-monitor/internal/report"
)

func writeGovernanceSummary(w Writer, d model.Report, mode SummaryMode) {
	summaryWrapStart(w, mode, "governance")
	w.WriteHTML(`<div class="gov-summary">`)

	if len(d.Proposals) > 0 {
		w.WriteHTML(`<div class="gov-summary__cards">`)
		limit := len(d.Proposals)
		if limit > 3 {
			limit = 3
		}
		for _, pr := range d.Proposals[:limit] {
			w.WriteHTML(`<div class="gov-summary__card">`)
			w.WriteHTML(fmt.Sprintf(`<span class="gov-summary__card-id">#%d</span>`, pr.ID))
			w.WriteHTML(fmt.Sprintf(`<span class="gov-summary__card-title">%s</span>`,
				html.EscapeString(report.Truncate(pr.Title, 36))))
			if pr.HasTally {
				w.WriteHTML(`<div class="gov-summary__tally">`)
				w.WriteHTML(fmt.Sprintf(`<span class="gov-summary__tally-yes" title="yes %s">Y</span>`, html.EscapeString(pr.TallyYes)))
				w.WriteHTML(fmt.Sprintf(`<span class="gov-summary__tally-no" title="no %s">N</span>`, html.EscapeString(pr.TallyNo)))
				w.WriteHTML(fmt.Sprintf(`<span class="gov-summary__tally-veto" title="veto %s">V</span>`, html.EscapeString(pr.TallyVeto)))
				w.WriteHTML(`</div>`)
			}
			w.WriteHTML(`</div>`)
		}
		w.WriteHTML(`</div>`)
	} else {
		w.WriteHTML(`<p class="gov-summary__empty">No active voting proposals</p>`)
	}

	w.WriteHTML(`<div class="gov-summary__pills">`)
	w.WriteHTML(fmt.Sprintf(`<span class="gov-summary__pill">Deposit period: <strong>%d</strong></span>`,
		len(d.DepositProposals)))
	upgrade := "none scheduled"
	if d.UpgradeName != "" && d.UpgradeName != "none" {
		upgrade = d.UpgradeName + " @ " + d.UpgradeHeight
		if d.BlocksLeft != "" {
			upgrade += " (" + d.BlocksLeft + " blocks)"
		}
	}
	w.WriteHTML(fmt.Sprintf(`<span class="gov-summary__pill">Upgrade: <strong>%s</strong></span>`, html.EscapeString(upgrade)))
	w.WriteHTML(fmt.Sprintf(`<span class="gov-summary__pill">IBC clients: <strong>%d</strong> · token pairs: <strong>%d</strong></span>`,
		d.IBCClients, len(d.TokenPairs)))
	w.WriteHTML(`</div></div>`)
	summaryWrapEnd(w, mode)
}

func writeGovernance(w Writer, d model.Report) {
	w.Section("6. GOVERNANCE")
	writeEmbeddedSectionIntro(w, "Active and deposit-period proposals, `x/gov` voting rules, scheduled upgrades, IBC clients, and ERC-20 token pairs.")
	writeGovernanceSummary(w, d, SummaryEmbedded)

	if len(d.Proposals) > 0 {
		w.Subsection(fmt.Sprintf("Active Proposals  (%d)", len(d.Proposals)))
		for _, pr := range d.Proposals {
			item := fmt.Sprintf("**#%d** %s  _(voting ends %s)_", pr.ID, report.Truncate(pr.Title, 40), pr.End)
			if pr.HasTally {
				item += fmt.Sprintf("\n  - yes %s  no %s  abstain %s  veto %s",
					pr.TallyYes, pr.TallyNo, pr.TallyAbstain, pr.TallyVeto)
			}
			w.ListItem(item)
		}
		w.BlankLine()
	}

	if len(d.DepositProposals) > 0 {
		w.Subsection(fmt.Sprintf("Deposit-Period Proposals  (%d)", len(d.DepositProposals)))
		for _, pr := range d.DepositProposals {
			w.ListItem(fmt.Sprintf("**#%d** %s  _(deposit ends %s)_", pr.ID, report.Truncate(pr.Title, 40), pr.End))
		}
		w.BlankLine()
	}

	if len(d.Proposals)+len(d.DepositProposals) == 0 {
		w.Em("No active proposals.")
	}

	w.WriteHTML(governanceDomainCardsHTML(d))

	if len(d.TokenPairs) > 0 {
		w.Subsection(fmt.Sprintf("Token Pairs  (%d)", len(d.TokenPairs)))
		w.WriteHTML(governanceTokenPairsHTML(d.TokenPairs))
	}
	writeSectionSources(w, ViewGovernance, d)
}
