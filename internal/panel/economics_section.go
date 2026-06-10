package panel

import (
	"fmt"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func writeEconomicsSummary(w Writer, d model.Report, mode SummaryMode) {
	summaryWrapStart(w, mode, "economics")
	w.WriteHTML(`<div class="eco-summary">`)
	w.WriteHTML(`<div class="economics-kpi-band">`)
	writeEconomicsKPIRows(w, d)
	w.WriteHTML(`</div>`)
	b := pmtPoolBadge(d)
	writeSummaryBadges(w, "eco-summary__badges", b)
	w.WriteHTML(fmt.Sprintf(
		`<p class="eco-summary__secondary">Bonded <strong>%.2f%%</strong> · inflation <strong>%.2f%%</strong></p>`,
		d.BondedPct, d.Inflation))
	w.WriteHTML(`</div>`)
	summaryWrapEnd(w, mode)
}

func writeEconomicsKPIRows(w Writer, d model.Report) {
	if total := RewardInPerBlockTotal(d); total != "—" {
		w.Row("reward in / block", total)
	}
	if bal := FeeCollectorBalance(d); bal != "" {
		check := economicsFeeCollectorCheck(d)
		val := bal
		if check == "cleared" {
			val = bal + "  _(cleared each block)_"
		} else if check == "not cleared?" {
			val = bal + "  _(stuck — check distribution)_"
		}
		w.Row("fee_collector", val)
	}
	if d.CommunityPool != "" {
		w.Row("community pool", d.CommunityPool)
	}
	if del := economicsUnclaimedDelegator(d); del != "" {
		w.Row("unclaimed delegator", del)
	}
	if comm := economicsUnclaimedCommission(d); comm != "" {
		w.Row("unclaimed commission", comm)
	}
	if d.PMTEnabled && d.PMTBalance != "" {
		val := d.PMTBalance
		if d.PMTRunway != "" {
			val += "  (" + d.PMTRunway + ")"
		}
		w.Row("PMT pool", val)
	}
}

func writeEconomicsOverview(w Writer, d model.Report) {
	writeEconomicsLedger(w, d)
}

func writeEconomicsLedger(w Writer, d model.Report) {
	rows := economicsLedgerRows(d)
	if len(rows) == 0 {
		return
	}
	w.Subsection("Block reward ledger")
	w.Hint("`In this block` → derived (per-block mint, PMT, fees); `Balance now` → module x/bank balances; `Check` → derived (fee_collector cleared, pool drift); `reward flow` → derived (BeginBlock via fee_collector and x/distribution, see table).")
	w.Table([]string{"Step", "Where", "In this block", "Balance now", "Check"}, rows)
}
