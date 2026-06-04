package markdown

import (
	"fmt"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
	"github.com/arkantos1482/cosmos-monitor/internal/report"
)

func writeValidators(m *mdWriter, d model.Report) {
	// ── 3. VALIDATOR SET ─────────────────────────────────────────────────────
	m.section("3. VALIDATOR SET")

	fmt.Fprintf(m.w, "_All validators on the chain — identity and P2P per validator, then stake and security tables._\n\n")

	m.subsection("Network (P2P)")
	m.hint("Per validator: `p2p dial` / `node ID` → CometBFT `/status` (this node) or `/net_info` (peers); `operator` / `consensus` → REST `GET /cosmos/staking/v1beta1/validators`.")
	for _, v := range d.Validators {
		hdr := "**" + v.Moniker + "**"
		if v.IsLocal {
			hdr += "  _(this node)_"
		}
		fmt.Fprintf(m.w, "\n%s\n\n", hdr)
		if v.Operator != "" {
			m.row("operator", "`"+v.Operator+"`")
		} else {
			m.row("operator", "—")
		}
		p2p := v.P2PDial
		if p2p == "" {
			p2p = "—  _(not in this node's `/net_info` peers)_"
		}
		m.row("p2p dial", "`"+p2p+"`")
		if v.NodeID != "" {
			m.row("node ID", "`"+v.NodeID+"`")
		} else {
			m.row("node ID", "—")
		}
		if v.ConsensusAddr != "" {
			m.row("consensus", "`"+v.ConsensusAddr+"`")
		} else {
			m.row("consensus", "—")
		}
	}
	fmt.Fprintln(m.w)

	m.subsection("Stake")
	m.hint("`vp%%`, `commission`, `status` → REST `GET /cosmos/staking/v1beta1/validators` (bonded, unbonding, unbonded).")
	fmt.Fprintf(m.w, "| moniker | vp%% | commission | status | local |\n")
	fmt.Fprintf(m.w, "|---------|-----|------------|--------|-------|\n")
	for _, v := range d.Validators {
		fmt.Fprintf(m.w, "| %s | %.1f%% | %.1f%% | %s | %s |\n",
			report.Truncate(v.Moniker, 14),
			v.VPFloat,
			v.CommissionFloat,
			v.Status,
			valLocalMark(v),
		)
	}
	fmt.Fprintln(m.w)

	m.subsection("Security")
	m.hint("`missed`, `tombstoned` → REST `GET /cosmos/slashing/v1beta1/signing_infos`; `jailed` → `x/staking` validators; `health` → derived (missed vs `min_signed_per_window` from slashing params).")
	fmt.Fprintf(m.w, "| moniker | missed | jailed | tombstoned | health | local |\n")
	fmt.Fprintf(m.w, "|---------|--------|--------|------------|--------|-------|\n")
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
		fmt.Fprintf(m.w, "| %s | %s | %s | %s | %s | %s |\n",
			report.Truncate(v.Moniker, 14),
			missed,
			jailed,
			tomb,
			health,
			valLocalMark(v),
		)
	}
	fmt.Fprintln(m.w)

	m.subsection("Summary")
	m.hint("`bonded` / `jailed` / `tombstoned` / `below min signed` → counts from §3 tables; `next proposer` → CometBFT `GET /validators` (highest `proposer_priority`).")
	m.row("bonded", fmt.Sprintf("%d", d.BondedCount))
	m.row("jailed", fmt.Sprintf("%d", d.JailedCount))
	m.row("tombstoned", fmt.Sprintf("%d", d.TombstonedCount))
	m.row("below min signed", fmt.Sprintf("%d", d.BelowThreshold))
	if d.NextProposer != "" {
		m.row("next proposer", d.NextProposer)
	}
}
