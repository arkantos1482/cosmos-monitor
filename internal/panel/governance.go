package panel

import (
	"fmt"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
	"github.com/arkantos1482/cosmos-monitor/internal/report"
)

func writeGovernance(w Writer, d model.Report) {
	w.Section("6. GOVERNANCE")

	if len(d.Proposals) > 0 {
		w.Subsection(fmt.Sprintf("Active Proposals  (%d)", len(d.Proposals)))
		w.Hint("`id`, `title`, `tally` → REST GET /cosmos/gov/v1beta1/proposals?proposal_status=2 (v1 fallback; per-proposal tally when available).")
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
		w.Hint("`id`, `title` → REST GET /cosmos/gov/v1beta1/proposals?proposal_status=1 (deposit period).")
		for _, pr := range d.DepositProposals {
			w.ListItem(fmt.Sprintf("**#%d** %s  _(deposit ends %s)_", pr.ID, report.Truncate(pr.Title, 40), pr.End))
		}
		w.BlankLine()
	}

	if len(d.Proposals)+len(d.DepositProposals) == 0 {
		w.Em("No active proposals.")
	}

	w.Subsection("Voting Params")
	w.Hint("`voting period` → REST GET /cosmos/gov/v1beta1/params/voting; `quorum`, `threshold`, `veto threshold` → REST GET …/params/tallying.")
	w.Row("voting period", d.VotingPeriod)
	w.Row("quorum", fmt.Sprintf("%.1f%%", d.Quorum))
	w.Row("threshold", fmt.Sprintf("%.1f%%", d.Threshold))
	if d.VetoThreshold > 0 {
		w.Row("veto threshold", fmt.Sprintf("%.1f%%", d.VetoThreshold))
	}

	w.Subsection("Upgrade")
	w.Hint("`name`, `target height` → REST GET /cosmos/upgrade/v1beta1/current_plan (plan null when none pending).")
	if d.UpgradeName == "" {
		w.Row("pending", "none")
	} else {
		w.Row("name", d.UpgradeName)
		w.Row("target height", d.UpgradeHeight)
		if d.BlocksLeft != "" {
			w.Row("blocks remaining", d.BlocksLeft)
		}
	}

	w.Subsection("IBC")
	w.Hint("`active clients` → derived (count of REST GET /ibc/core/client/v1/client_states).")
	w.Row("active clients", fmt.Sprintf("%d", d.IBCClients))

	w.Subsection(fmt.Sprintf("Token Pairs  (%d)", len(d.TokenPairs)))
	w.Hint("each row → REST GET /cosmos/evm/erc20/v1/token_pairs (denom, erc20_address, enabled).")
	if len(d.TokenPairs) == 0 {
		w.WriteString("none registered\n\n")
	}
	for _, tp := range d.TokenPairs {
		enabled := "yes"
		if !tp.Enabled {
			enabled = "no"
		}
		w.ListItem(fmt.Sprintf("`%s`  `%s`  enabled: %s", tp.Denom, tp.ERC20, enabled))
	}
}
