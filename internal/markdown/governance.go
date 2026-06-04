package markdown

import (
	"fmt"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
	"github.com/arkantos1482/cosmos-monitor/internal/report"
)

func writeGovernance(m *mdWriter, d model.Report) {
	// ── 6. GOVERNANCE ────────────────────────────────────────────────────────
	m.section("6. GOVERNANCE")

	if len(d.Proposals) > 0 {
		m.subsection(fmt.Sprintf("Active Proposals  (%d)", len(d.Proposals)))
		m.hint("`GET /cosmos/gov/v1beta1/proposals?proposal_status=2` (v1 fallback if empty); tallies from per-proposal tally queries when available.")
		for _, pr := range d.Proposals {
			fmt.Fprintf(m.w, "- **#%d** %s  _(voting ends %s)_\n", pr.ID, report.Truncate(pr.Title, 40), pr.End)
			if pr.HasTally {
				fmt.Fprintf(m.w, "  - yes %s  no %s  abstain %s  veto %s\n",
					pr.TallyYes, pr.TallyNo, pr.TallyAbstain, pr.TallyVeto)
			}
		}
		fmt.Fprintln(m.w)
	}

	if len(d.DepositProposals) > 0 {
		m.subsection(fmt.Sprintf("Deposit-Period Proposals  (%d)", len(d.DepositProposals)))
		m.hint("`GET /cosmos/gov/v1beta1/proposals?proposal_status=1` (deposit period).")
		for _, pr := range d.DepositProposals {
			fmt.Fprintf(m.w, "- **#%d** %s  _(deposit ends %s)_\n", pr.ID, report.Truncate(pr.Title, 40), pr.End)
		}
		fmt.Fprintln(m.w)
	}

	if len(d.Proposals)+len(d.DepositProposals) == 0 {
		fmt.Fprintf(m.w, "_No active proposals._\n\n")
	}

	m.subsection("Voting Params")
	m.hint("`voting period` → `GET /cosmos/gov/v1beta1/params/voting`; `quorum`, `threshold`, `veto threshold` → `…/params/tallying`.")
	m.row("voting period", d.VotingPeriod)
	m.row("quorum", fmt.Sprintf("%.1f%%", d.Quorum))
	m.row("threshold", fmt.Sprintf("%.1f%%", d.Threshold))
	if d.VetoThreshold > 0 {
		m.row("veto threshold", fmt.Sprintf("%.1f%%", d.VetoThreshold))
	}

	m.subsection("Upgrade")
	m.hint("`name`, `target height` → `GET /cosmos/upgrade/v1beta1/current_plan` (`plan` null when none pending).")
	if d.UpgradeName == "" {
		m.row("pending", "none")
	} else {
		m.row("name", d.UpgradeName)
		m.row("target height", d.UpgradeHeight)
		if d.BlocksLeft != "" {
			m.row("blocks remaining", d.BlocksLeft)
		}
	}

	m.subsection("IBC")
	m.hint("`active clients` → count of `GET /ibc/core/client/v1/client_states`.")
	m.row("active clients", fmt.Sprintf("%d", d.IBCClients))

	m.subsection(fmt.Sprintf("Token Pairs  (%d)", len(d.TokenPairs)))
	m.hint("Each row → `GET /cosmos/evm/erc20/v1/token_pairs` (`denom`, `erc20_address`, `enabled`).")
	if len(d.TokenPairs) == 0 {
		fmt.Fprintf(m.w, "none registered\n\n")
	}
	for _, tp := range d.TokenPairs {
		enabled := "yes"
		if !tp.Enabled {
			enabled = "no"
		}
		fmt.Fprintf(m.w, "- `%s`  `%s`  enabled: %s\n", tp.Denom, tp.ERC20, enabled)
	}
}
