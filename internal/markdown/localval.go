package markdown

import (
	"fmt"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
	"github.com/arkantos1482/cosmos-monitor/internal/report"
)

func writeLocalValidator(m *mdWriter, d model.Report) {
	// ── 4. THIS VALIDATOR ──────────────────────────────────────────────────────
	m.section("4. THIS VALIDATOR")

	fmt.Fprintf(m.w, "_Staking and rewards for this machine's validator — matched via `/status` consensus address. Node identity is in §2._\n\n")

	lv := d.Local
	if !lv.IsValidator {
		m.row("role", lv.SigningStatus)
		m.row("moniker", d.Moniker)
	} else {
		m.subsection("Operator")
		m.hint("`operator address`, `moniker` → matched local validator from `x/staking` (consensus address from `/status` `validator_info`).")
		if lv.OperatorAddr != "" {
			m.row("operator address", lv.OperatorAddr+"  _(staking / rewards — `evmd query/distribution` use this)_")
		}
		m.row("moniker", lv.Moniker)

		m.subsection("Staking")
		m.hint("`status`, `jailed`, `voting power`, `commission` → `x/staking` validators; `outstanding rewards` / `commission earned` → `x/distribution` per-valoper (`…/outstanding_rewards`, `…/commission`).")
		m.row("status", lv.Status)
		if lv.Jailed {
			m.row("jailed", "yes")
		}
		if lv.Tombstoned {
			m.row("tombstoned", "YES")
		}
		m.row("voting power", fmt.Sprintf("%s  (%.1f%% of bonded stake)", lv.VotingPower, lv.VPPercent))
		m.row("commission", fmt.Sprintf("%.1f%%  _(your cut of delegator rewards)_", lv.Commission))
		if lv.Outstanding != "" {
			m.row("outstanding rewards", lv.Outstanding+"  _(total rewards not yet withdrawn — x/distribution)_")
		} else {
			m.row("outstanding rewards", "–")
		}
		if lv.CommissionEarned != "" {
			m.row("commission earned", lv.CommissionEarned+"  _(validator commission, unclaimed — x/distribution)_")
		} else {
			m.row("commission earned", "–")
		}

		m.subsection("Block Signing")
		m.hint("`signing health`, `missed / window` → `x/slashing` signing_infos + params; `proposer` / `proposer priority` → CometBFT `GET /validators`.")
		m.row("signing health", lv.SigningStatus)
		if d.SlashWindow != "" && d.SlashWindow != "0" {
			m.row("missed / window", fmt.Sprintf("%d / %s blocks  (max allowed: %d)", lv.Missed, d.SlashWindow, lv.MaxMissed))
		}
		if lv.IsNextProposer {
			m.row("proposer", "**next block proposer**")
		} else if d.NextProposer != "" {
			m.row("proposer", "not next  _(next: "+d.NextProposer+")_")
		}
		if lv.ProposerPriority != 0 {
			m.row("proposer priority", report.FormatInt(lv.ProposerPriority))
		}
	}
}
