package panel

import (
	"fmt"
	"html"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
	"github.com/arkantos1482/cosmos-monitor/internal/report"
)

func writeValidatorP2PSummaryBody(w Writer, d model.Report) {
	w.WriteHTML(`<div class="val-summary val-summary--p2p">`)
	w.WriteHTML(`<p class="val-summary__heading">Validator set</p>`)
	w.WriteHTML(`<div class="val-summary__kpis">`)
	if d.BondedCount > 0 {
		writeValSummaryKPI(w, "active set", fmt.Sprintf("%d", d.BondedCount), "")
	}
	if d.BondedPct > 0 {
		writeValSummaryKPI(w, "bonded", fmt.Sprintf("%.1f%%", d.BondedPct), "")
	}
	if unhealthy := validatorUnhealthyCount(d); unhealthy > 0 {
		writeValSummaryKPI(w, "unhealthy", fmt.Sprintf("%d", unhealthy), "warn")
	}
	w.WriteHTML(`</div>`)
	if len(d.Validators) > 0 {
		w.WriteHTML(`<div class="val-summary__chips">`)
		for _, v := range d.Validators {
			cls := chipClass(v)
			if v.IsLocal {
				cls += " val-summary__chip--local"
			}
			vp := ""
			if v.VPFloat > 0 {
				vp = fmt.Sprintf(` <span class="val-summary__chip-vp">%.0f%%</span>`, v.VPFloat)
			}
			title := v.Status
			if v.Jailed {
				title = "jailed"
			} else if v.Tombstoned {
				title = "tombstoned"
			}
			w.WriteHTML(fmt.Sprintf(
				`<span class="val-summary__chip%s" title="%s">%s%s</span>`,
				cls, html.EscapeString(title), html.EscapeString(report.Truncate(v.Moniker, 14)), vp))
		}
		w.WriteHTML(`</div>`)
	}
	w.WriteHTML(`</div>`)
}

func validatorUnhealthyCount(d model.Report) int {
	return d.JailedCount + d.BelowThreshold + d.TombstonedCount
}

func writeValSummaryKPI(w Writer, label, value, tone string) {
	if value == "" {
		return
	}
	valCls := "val-summary__kpi-val"
	if tone != "" {
		valCls += " val-summary__kpi-val--" + tone
	}
	w.WriteHTML(fmt.Sprintf(
		`<div class="val-summary__kpi"><span class="val-summary__kpi-label">%s</span>`+
			`<span class="%s">%s</span></div>`,
		html.EscapeString(label), valCls, html.EscapeString(value)))
}

func chipClass(v model.Validator) string {
	if v.Jailed || v.Tombstoned || v.MissedHigh {
		return " val-summary__chip--warn"
	}
	return ""
}

func writeValidatorP2PNetwork(w Writer, d model.Report) {
	w.Layer("Validator set")
	w.Subsection("Network (P2P)")
	p2pRows := make([][]string, 0, len(d.Validators))
	for _, v := range d.Validators {
		cons := v.ConsensusBech32
		if cons == "" {
			cons = v.ConsensusAddr
		}
		p2pRows = append(p2pRows, []string{
			report.Truncate(v.Moniker, 14),
			identityCell(v.P2PDial),
			identityCell(v.NodeID),
			identityCell(cons),
		})
	}
	writeValidatorSetTable(w, []string{"moniker", "p2p dial", "node ID", "consensus"}, p2pRows, d.Validators)
}

func identityCell(s string) string {
	if s == "" {
		return "—"
	}
	return "`" + s + "`"
}
