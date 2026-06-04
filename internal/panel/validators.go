package panel

import (
	"fmt"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
	"github.com/arkantos1482/cosmos-monitor/internal/report"
)

func writeValidators(w Writer, d model.Report) {
	w.Section("3. VALIDATOR SET")
	w.Em("All validators on the chain — identity and P2P per validator, then stake and security tables.")

	w.Subsection("Network (P2P)")
	w.Hint("Per validator: `p2p dial` / `node ID` → CometBFT `/status` (this node) or `/net_info` (peers); `operator` / `consensus` → REST `GET /cosmos/staking/v1beta1/validators`.")
	for _, v := range d.Validators {
		hdr := "**" + v.Moniker + "**"
		if v.IsLocal {
			hdr += "  _(this node)_"
		}
		w.ValidatorHeader(hdr)
		if v.Operator != "" {
			w.Row("operator", "`"+v.Operator+"`")
		} else {
			w.Row("operator", "—")
		}
		p2p := v.P2PDial
		if p2p == "" {
			p2p = "—  _(not in this node's `/net_info` peers)_"
		}
		w.Row("p2p dial", "`"+p2p+"`")
		if v.NodeID != "" {
			w.Row("node ID", "`"+v.NodeID+"`")
		} else {
			w.Row("node ID", "—")
		}
		if v.ConsensusAddr != "" {
			w.Row("consensus", "`"+v.ConsensusAddr+"`")
		} else {
			w.Row("consensus", "—")
		}
	}
	w.BlankLine()

	w.Subsection("Stake")
	w.Hint("`vp%%`, `commission`, `status` → REST `GET /cosmos/staking/v1beta1/validators` (bonded, unbonding, unbonded).")
	stakeRows := make([][]string, 0, len(d.Validators))
	for _, v := range d.Validators {
		stakeRows = append(stakeRows, []string{
			report.Truncate(v.Moniker, 14),
			fmt.Sprintf("%.1f%%", v.VPFloat),
			fmt.Sprintf("%.1f%%", v.CommissionFloat),
			v.Status,
			valLocalMark(v),
		})
	}
	w.Table([]string{"moniker", "vp%", "commission", "status", "local"}, stakeRows)

	w.Subsection("Security")
	w.Hint("`missed`, `tombstoned` → REST `GET /cosmos/slashing/v1beta1/signing_infos`; `jailed` → `x/staking` validators; `health` → derived (missed vs `min_signed_per_window` from slashing params).")
	secRows := make([][]string, 0, len(d.Validators))
	for _, v := range d.Validators {
		missed := fmt.Sprintf("%d", v.Missed)
		health := "ok"
		if v.Tombstoned {
			health = "tombstoned"
		} else if v.Jailed {
			health = "jailed"
		} else if v.MissedHigh {
			health = "⚠ below min signed"
			missed += " ⚠"
		} else if v.Missed > 0 {
			health = "ok (some misses)"
		}
		jailed := ""
		if v.Jailed {
			jailed = "yes"
		}
		tomb := ""
		if v.Tombstoned {
			tomb = "yes"
		}
		secRows = append(secRows, []string{
			report.Truncate(v.Moniker, 14),
			missed,
			jailed,
			tomb,
			health,
			valLocalMark(v),
		})
	}
	w.Table([]string{"moniker", "missed", "jailed", "tombstoned", "health", "local"}, secRows)

	w.Subsection("Summary")
	w.Hint("`bonded` / `jailed` / `tombstoned` / `below min signed` → counts from §3 tables; `next proposer` → CometBFT `GET /validators` (highest `proposer_priority`).")
	w.Row("bonded", fmt.Sprintf("%d", d.BondedCount))
	w.Row("jailed", fmt.Sprintf("%d", d.JailedCount))
	w.Row("tombstoned", fmt.Sprintf("%d", d.TombstonedCount))
	w.Row("below min signed", fmt.Sprintf("%d", d.BelowThreshold))
	if d.NextProposer != "" {
		w.Row("next proposer", d.NextProposer)
	}
}
