package panel

import (
	"fmt"

	"github.com/arkantos1482/cosmos-monitor/internal/fetch"
	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func writeEconomicsOverview(w Writer, d model.Report) {
	w.Subsection("Money flow (live balances)")
	w.Hint("On-chain values from REST: module `bank` balances, `x/distribution` outstanding/commission sums, and PMT/mint params. Refreshes with the dashboard.")

	writeEconomicsSourceTable(w, d)
	writeEconomicsModuleTable(w, d)
	writeEconomicsSplitTable(w, d)
	writeEconomicsNetworkTable(w, d)
	if d.Local.IsValidator {
		writeEconomicsLocalTable(w, d)
	}
}

func writeEconomicsSourceTable(w Writer, d model.Report) {
	var rows [][]string
	if d.InflationPerBlock != "" {
		rows = append(rows, []string{
			"x/mint inflation",
			d.InflationPerBlock,
			d.InflationPerDay,
			fmt.Sprintf("%.2f%% inflation  _(annual provisions ÷ blocks/year)_", d.Inflation),
		})
	} else if d.Inflation == 0 {
		rows = append(rows, []string{"x/mint inflation", "0", "0", "inactive  _(0% inflation)_"})
	}
	if d.PMTEnabled && d.PMTRate != "" {
		rows = append(rows, []string{
			"x/pmtrewards pool",
			d.PMTRate,
			d.PMTDailyEmit,
			"mint hook → fee_collector  _(pool " + d.PMTBalance + ")_",
		})
	}
	if d.LastBlockFees != "" {
		rows = append(rows, []string{
			"tx fees (parent block)",
			d.LastBlockFees,
			"—",
			"estimate from gas used × base fee  _(feemarket)_",
		})
	} else if d.MempoolTxs > 0 || d.GasPrice != "" {
		rows = append(rows, []string{
			"tx fees (mempool)",
			"—",
			"—",
			fmt.Sprintf("mempool %d · gas %s  _(per-block fee not available)_", d.MempoolTxs, fetch.FormatFeeAmount(d.GasPrice, economicsDenom(d))),
		})
	}
	if len(rows) == 0 {
		return
	}
	w.Table([]string{"Source", "Per block", "Per day (est.)", "Note"}, rows)
}

func writeEconomicsModuleTable(w Writer, d model.Report) {
	if len(d.ModuleAccounts) == 0 {
		return
	}
	var rows [][]string
	for _, m := range d.ModuleAccounts {
		addr := m.Address
		if addr == "" {
			addr = "—"
		}
		rows = append(rows, []string{m.Name, m.Balance, addr, m.Role})
	}
	w.Table([]string{"Module account", "Balance", "Address", "Role"}, rows)
}

func writeEconomicsSplitTable(w Writer, d model.Report) {
	tax, valPool, op, del, unit, ok := economicsPerBlockSplit(d)
	if !ok && d.CommunityTax == "" {
		return
	}
	var rows [][]string
	if ok {
		rows = append(rows, []string{
			"community tax → pool",
			economicsFormatPerBlock(tax, unit),
			economicsFormatPerDay(d, tax, unit),
			d.CommunityTax + " of block rewards",
		})
		rows = append(rows, []string{
			"validator pool (all vals)",
			economicsFormatPerBlock(valPool, unit),
			economicsFormatPerDay(d, valPool, unit),
			fmt.Sprintf("split across %d bonded validators (equal-share est.)", max(1, d.BondedCount)),
		})
		if pct, has := economicsCommissionPct(d); has {
			rows = append(rows, []string{
				"→ operator commission (avg/val)",
				economicsFormatPerBlock(op, unit),
				economicsFormatPerDay(d, op, unit),
				fmt.Sprintf("%.1f%% commission (network avg)", pct),
			})
			rows = append(rows, []string{
				"→ delegator rewards (avg/val)",
				economicsFormatPerBlock(del, unit),
				economicsFormatPerDay(d, del, unit),
				fmt.Sprintf("%.1f%% of validator share", 100-pct),
			})
		}
	} else {
		rows = append(rows, []string{
			"community tax",
			"—",
			"—",
			d.CommunityTax,
		})
	}
	w.Table([]string{"Distribution split", "Per block (est.)", "Per day (est.)", "Basis"}, rows)
}

func writeEconomicsNetworkTable(w Writer, d model.Report) {
	var rows [][]string
	rows = append(rows, []string{
		"community pool",
		d.CommunityPool,
		"governance treasury  _(distribution community_pool)_",
	})
	if del := economicsUnclaimedDelegator(d); del != "" {
		rows = append(rows, []string{
			"unclaimed delegator rewards",
			del,
			"sum outstanding_rewards (all validators)",
		})
	}
	if comm := economicsUnclaimedCommission(d); comm != "" {
		rows = append(rows, []string{
			"unclaimed validator commission",
			comm,
			"sum accumulated commission (all validators)",
		})
	}
	if total := economicsUnclaimedTotal(d); total != "" {
		rows = append(rows, []string{
			"total unclaimed staking rewards",
			total,
			splitOutstandingNote(d),
		})
	}
	if d.BondedAmt != "" {
		rows = append(rows, []string{
			"bonded stake (staking pool)",
			d.BondedAmt,
			fmt.Sprintf("%.2f%% of supply · goal %.0f%%", d.BondedPct, d.GoalBonded),
		})
	}
	w.Table([]string{"Network total", "Amount", "Note"}, rows)
}

func splitOutstandingNote(d model.Report) string {
	if _, suffix := splitOutstandingSuffix(d.TotalOutstanding); suffix != "" {
		return suffix
	}
	if d.BondedCount > 0 {
		return fmt.Sprintf("%d validators", d.BondedCount)
	}
	return ""
}

func writeEconomicsLocalTable(w Writer, d model.Report) {
	lv := d.Local
	var rows [][]string
	if lv.Moniker != "" {
		rows = append(rows, []string{"moniker", lv.Moniker, ""})
	}
	if lv.OperatorAddr != "" {
		rows = append(rows, []string{"operator", lv.OperatorAddr, ""})
	}
	if lv.VotingPower != "" {
		note := economicsLocalVPShare(d)
		rows = append(rows, []string{"voting power", lv.VotingPower, note})
	}
	if lv.Commission > 0 {
		rows = append(rows, []string{
			"commission rate",
			fmt.Sprintf("%.2f%%", lv.Commission),
			"operator share of this validator's rewards",
		})
	}
	if lv.CommissionEarned != "" {
		rows = append(rows, []string{
			"unclaimed commission",
			lv.CommissionEarned,
			"GET …/distribution/v1beta1/validators/{op}/commission",
		})
	}
	if lv.Outstanding != "" {
		rows = append(rows, []string{
			"outstanding to delegators",
			lv.Outstanding,
			"GET …/validators/{op}/outstanding_rewards",
		})
	}
	if _, valPool, _, _, unit, ok := economicsPerBlockSplit(d); ok && lv.VPPercent > 0 {
		vp := lv.VPPercent / 100
		localShare := valPool * vp
		localOp := localShare * lv.Commission / 100
		localDel := localShare - localOp
		rows = append(rows, []string{
			"est. commission / block",
			economicsFormatPerBlock(localOp, unit),
			fmt.Sprintf("%.2f%% VP · %.2f%% commission", lv.VPPercent, lv.Commission),
		})
		rows = append(rows, []string{
			"est. delegator share / block",
			economicsFormatPerBlock(localDel, unit),
			"estimate from PMTRate — compare to outstanding balances",
		})
	}
	w.Table([]string{"This validator", "Value", "Note"}, rows)
}
