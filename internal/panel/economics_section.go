package panel

import (
	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func writeEconomicsOverview(w Writer, d model.Report) {
	writeEconomicsAtAGlance(w, d)
	writeEconomicsLedger(w, d)
}

func writeEconomicsAtAGlance(w Writer, d model.Report) {
	w.Subsection("At a glance")
	w.Hint("`reward in / block` → derived (PMT + inflation + last-block fees, ledger); `fee_collector` → module x/bank (clears each BeginBlock); `community pool` → REST GET /cosmos/distribution/v1beta1/community_pool; `unclaimed delegator`, `unclaimed commission` → module x/distribution outstanding totals; `PMT pool` → module x/pmtrewards pool account.")
	w.WriteHTML(`<div class="economics-kpi-band">`)

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
	w.WriteHTML(`</div>`)
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
