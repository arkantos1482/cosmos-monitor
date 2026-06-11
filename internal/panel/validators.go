package panel

import (
	"fmt"
	"html"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
	"github.com/arkantos1482/cosmos-monitor/internal/report"
)

func writeValidatorsSummary(w Writer, d model.Report, mode SummaryMode) {
	summaryWrapStart(w, mode, "validators")
	w.WriteHTML(`<div class="val-summary">`)
	w.WriteHTML(`<div class="val-summary__kpis">`)
	for _, kpi := range []struct{ label, val string }{
		{"bonded", fmt.Sprintf("%d", d.BondedCount)},
		{"jailed", fmt.Sprintf("%d", d.JailedCount)},
		{"tombstoned", fmt.Sprintf("%d", d.TombstonedCount)},
		{"below min signed", fmt.Sprintf("%d", d.BelowThreshold)},
	} {
		w.WriteHTML(fmt.Sprintf(
			`<div class="val-summary__kpi"><span class="val-summary__kpi-label">%s</span>`+
				`<span class="val-summary__kpi-val">%s</span></div>`,
			html.EscapeString(kpi.label), html.EscapeString(kpi.val)))
	}
	w.WriteHTML(`</div>`)
	if d.NextProposer != "" {
		w.WriteHTML(fmt.Sprintf(
			`<p class="val-summary__proposer">Next proposer: <strong>%s</strong></p>`,
			html.EscapeString(d.NextProposer)))
	}
	if d.JailedCount > 0 || d.BelowThreshold > 0 {
		var alerts []string
		if d.JailedCount > 0 {
			alerts = append(alerts, fmt.Sprintf("%d jailed", d.JailedCount))
		}
		if d.BelowThreshold > 0 {
			alerts = append(alerts, fmt.Sprintf("%d below min signed", d.BelowThreshold))
		}
		w.WriteHTML(fmt.Sprintf(`<p class="val-summary__alert">⚠ %s</p>`, html.EscapeString(alerts[0])))
		if len(alerts) > 1 {
			w.WriteHTML(fmt.Sprintf(`<p class="val-summary__alert">⚠ %s</p>`, html.EscapeString(alerts[1])))
		}
	}
	if len(d.Validators) > 0 {
		w.WriteHTML(`<div class="val-summary__chips">`)
		for _, v := range d.Validators {
			w.WriteHTML(fmt.Sprintf(
				`<span class="val-summary__chip%s">%s <span class="val-summary__chip-vp">%.1f%%</span></span>`,
				chipClass(v), html.EscapeString(report.Truncate(v.Moniker, 12)), v.VPFloat))
		}
		w.WriteHTML(`</div>`)
	}
	w.WriteHTML(`</div>`)
	summaryWrapEnd(w, mode)
}

func chipClass(v model.Validator) string {
	if v.Jailed || v.Tombstoned || v.MissedHigh {
		return " val-summary__chip--warn"
	}
	return ""
}

func writeValidators(w Writer, d model.Report) {
	w.Section("1. VALIDATOR SET")
	writeValidatorsSummary(w, d, SummaryEmbedded)
	w.Em("Chain-wide validator set — summary counts, stake and slashing tables, then P2P identity per validator.")

	w.Subsection("Stake")
	w.Hint("`vp%%`, `commission`, `status` → REST GET /cosmos/staking/v1beta1/validators (bonded, unbonding, unbonded).")
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

	w.Subsection("Slashing")
	w.Hint("`missed`, `tombstoned` → REST GET /cosmos/slashing/v1beta1/signing_infos; `jailed` → module x/staking validators; `health` → derived (missed vs min_signed_per_window from slashing params).")
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

	w.Subsection("Network (P2P)")
	w.Hint("`p2p dial`, `node ID` → CometBFT GET /status (local) or GET /net_info (peers); `operator`, `consensus` → REST GET /cosmos/staking/v1beta1/validators.")
	p2pRows := make([][]string, 0, len(d.Validators))
	for _, v := range d.Validators {
		cons := v.ConsensusBech32
		if cons == "" {
			cons = v.ConsensusAddr
		}
		p2pRows = append(p2pRows, []string{
			report.Truncate(v.Moniker, 14),
			identityCell(v.Operator),
			identityCell(v.P2PDial),
			identityCell(v.NodeID),
			identityCell(cons),
			valLocalMark(v),
		})
	}
	w.Table([]string{"moniker", "operator", "p2p dial", "node ID", "consensus", "local"}, p2pRows)
	w.BlankLine()
}

func identityCell(s string) string {
	if s == "" {
		return "—"
	}
	return "`" + s + "`"
}
