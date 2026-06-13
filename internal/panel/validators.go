package panel

import (
	"fmt"
	"html"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
	"github.com/arkantos1482/cosmos-monitor/internal/report"
)

func writeValidatorP2PSummaryBody(w Writer, d model.Report) {
	w.WriteHTML(`<div class="val-summary val-summary--p2p">`)
	if n := len(d.Validators); n > 0 {
		w.WriteHTML(fmt.Sprintf(`<p class="val-summary__count">%d validators</p>`, n))
	}
	if len(d.Validators) > 0 {
		w.WriteHTML(`<div class="val-summary__chips">`)
		for _, v := range d.Validators {
			w.WriteHTML(fmt.Sprintf(
				`<span class="val-summary__chip%s">%s</span>`,
				chipClass(v), html.EscapeString(report.Truncate(v.Moniker, 14))))
		}
		w.WriteHTML(`</div>`)
	}
	w.WriteHTML(`</div>`)
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
