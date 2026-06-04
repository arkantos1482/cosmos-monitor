package panel

import (
	"fmt"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
	"github.com/arkantos1482/cosmos-monitor/internal/report"
)

func writeLocalValidator(w Writer, d model.Report) {
	w.Section("4. THIS VALIDATOR")
	w.Em("Staking and rewards for this machine's validator — matched via `/status` consensus address. Node identity is in §2.")

	lv := d.Local
	if !lv.IsValidator {
		w.Row("role", lv.SigningStatus)
		w.Row("moniker", d.Moniker)
	} else {
		w.Subsection("Operator")
		w.Hint("`operator address`, `moniker` → matched local validator from `x/staking` (consensus address from `/status` `validator_info`).")
		if lv.OperatorAddr != "" {
			w.Row("operator address", lv.OperatorAddr+"  _(staking / rewards — `evmd query/distribution` use this)_")
		}
		w.Row("moniker", lv.Moniker)

		w.Subsection("Staking")
		w.Hint("`status`, `jailed`, `voting power`, `commission` → `x/staking` validators; `outstanding rewards` / `commission earned` → `x/distribution` per-valoper (`…/outstanding_rewards`, `…/commission`).")
		w.Row("status", lv.Status)
		if lv.Jailed {
			w.Row("jailed", "yes")
		}
		if lv.Tombstoned {
			w.Row("tombstoned", "YES")
		}
		w.Row("voting power", fmt.Sprintf("%s  (%.1f%% of bonded stake)", lv.VotingPower, lv.VPPercent))
		w.Row("commission", fmt.Sprintf("%.1f%%  _(your cut of delegator rewards)_", lv.Commission))
		if lv.Outstanding != "" {
			w.Row("outstanding rewards", lv.Outstanding+"  _(total rewards not yet withdrawn — x/distribution)_")
		} else {
			w.Row("outstanding rewards", "–")
		}
		if lv.CommissionEarned != "" {
			w.Row("commission earned", lv.CommissionEarned+"  _(validator commission, unclaimed — x/distribution)_")
		} else {
			w.Row("commission earned", "–")
		}

		w.Subsection("Block Signing")
		w.Hint("`signing health`, `missed / window` → `x/slashing` signing_infos + params; `proposer` / `proposer priority` → CometBFT `GET /validators`.")
		w.Row("signing health", lv.SigningStatus)
		if d.SlashWindow != "" && d.SlashWindow != "0" {
			w.Row("missed / window", fmt.Sprintf("%d / %s blocks  (max allowed: %d)", lv.Missed, d.SlashWindow, lv.MaxMissed))
		}
		if lv.IsNextProposer {
			w.Row("proposer", "**next block proposer**")
		} else if d.NextProposer != "" {
			w.Row("proposer", "not next  _(next: "+d.NextProposer+")_")
		}
		if lv.ProposerPriority != 0 {
			w.Row("proposer priority", report.FormatInt(lv.ProposerPriority))
		}
	}
}
