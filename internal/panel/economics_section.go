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
	w.Hint("Live REST balances and per-block rates. `fee_collector` should clear each BeginBlock after distribution.")

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
	if d.Local.IsValidator && d.Local.CommissionEarned != "" {
		w.Row("your commission", d.Local.CommissionEarned)
	}
}

func writeEconomicsLedger(w Writer, d model.Report) {
	rows := economicsLedgerRows(d)
	if len(rows) == 0 {
		return
	}
	w.Subsection("Block reward ledger")
	w.Hint("Follows BeginBlock: sources → `fee_collector` → community tax + validator pool → operators and delegators. **In this block** = rates/estimates; **Balance now** = on-chain balances.")
	w.Table([]string{"Step", "Where", "In this block", "Balance now", "Check"}, rows)
}
