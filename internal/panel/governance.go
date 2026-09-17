package panel

import (
	"fmt"
	"html"
	"strings"

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
			title := pr.Title
			if title == "" {
				title = pr.Messages
			}
			w.WriteHTML(`<div class="gov-summary__card">`)
			w.WriteHTML(fmt.Sprintf(`<span class="gov-summary__card-id">#%d</span>`, pr.ID))
			if pr.Expedited {
				w.WriteHTML(`<span class="gov-proposal__badge">expedited</span>`)
			}
			w.WriteHTML(fmt.Sprintf(`<span class="gov-summary__card-title">%s</span>`,
				html.EscapeString(report.Truncate(title, 48))))
			if pr.Summary != "" {
				w.WriteHTML(fmt.Sprintf(`<p class="gov-summary__card-body">%s</p>`,
					html.EscapeString(report.Truncate(pr.Summary, 90))))
			}
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
		w.WriteHTML(governanceProposalsHTML(d.Proposals, "voting ends"))
	}

	if len(d.DepositProposals) > 0 {
		w.Subsection(fmt.Sprintf("Deposit-Period Proposals  (%d)", len(d.DepositProposals)))
		w.WriteHTML(governanceProposalsHTML(d.DepositProposals, "deposit ends"))
	}

	if len(d.Proposals)+len(d.DepositProposals) == 0 {
		w.Em("No active proposals.")
	}

	if len(d.RecentProposals) > 0 {
		w.Subsection(fmt.Sprintf("Recent Proposals  (%d)", len(d.RecentProposals)))
		rows := make([][]string, 0, len(d.RecentProposals))
		for _, pr := range d.RecentProposals {
			kind := "standard"
			if pr.Expedited {
				kind = "expedited"
			}
			title := pr.Title
			if title == "" {
				title = pr.Messages
			}
			rows = append(rows, []string{
				fmt.Sprintf("#%d", pr.ID),
				title,
				pr.Status,
				kind,
				pr.End,
			})
		}
		w.Table([]string{"ID", "Title", "Status", "Kind", "Ended"}, rows)
	}

	w.WriteHTML(governanceDomainCardsHTML(d))

	if len(d.TokenPairs) > 0 {
		w.Subsection(fmt.Sprintf("Token Pairs  (%d)", len(d.TokenPairs)))
		w.WriteHTML(governanceTokenPairsHTML(d.TokenPairs))
	}
	writeSectionSources(w, ViewGovernance, d)
}

func governanceProposalsHTML(proposals []model.Proposal, endLabel string) string {
	var b strings.Builder
	b.WriteString(`<div class="gov-proposals">`)
	for _, pr := range proposals {
		b.WriteString(governanceProposalHTML(pr, endLabel))
	}
	b.WriteString(`</div>`)
	return b.String()
}

func governanceProposalHTML(pr model.Proposal, endLabel string) string {
	title := pr.Title
	if title == "" {
		title = "Untitled proposal"
	}
	var b strings.Builder
	b.WriteString(`<article class="gov-proposal">`)
	b.WriteString(`<div class="gov-proposal__head">`)
	fmt.Fprintf(&b, `<span class="gov-summary__card-id">#%d</span>`, pr.ID)
	if pr.Expedited {
		b.WriteString(`<span class="gov-proposal__badge">expedited</span>`)
	}
	fmt.Fprintf(&b, `<h3 class="gov-proposal__title">%s</h3>`, html.EscapeString(title))
	b.WriteString(`</div>`)
	var meta []string
	if pr.Messages != "" {
		meta = append(meta, pr.Messages)
	}
	if pr.End != "" {
		meta = append(meta, endLabel+" "+pr.End)
	}
	if len(meta) > 0 {
		fmt.Fprintf(&b, `<p class="gov-proposal__meta">%s</p>`, html.EscapeString(strings.Join(meta, " · ")))
	}
	if pr.Summary != "" {
		fmt.Fprintf(&b, `<p class="gov-proposal__summary">%s</p>`, html.EscapeString(pr.Summary))
	}
	if pr.HasTally {
		b.WriteString(`<div class="gov-summary__tally">`)
		fmt.Fprintf(&b, `<span class="gov-summary__tally-yes" title="yes">yes %s</span>`, html.EscapeString(pr.TallyYes))
		fmt.Fprintf(&b, `<span class="gov-summary__tally-no" title="no">no %s</span>`, html.EscapeString(pr.TallyNo))
		fmt.Fprintf(&b, `<span class="gov-summary__tally-veto" title="veto">veto %s</span>`, html.EscapeString(pr.TallyVeto))
		fmt.Fprintf(&b, `<span class="gov-summary__tally-abstain" title="abstain">abstain %s</span>`, html.EscapeString(pr.TallyAbstain))
		b.WriteString(`</div>`)
	}
	b.WriteString(`</article>`)
	return b.String()
}
