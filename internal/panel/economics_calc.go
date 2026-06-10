package panel

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/arkantos1482/cosmos-monitor/internal/fetch"
	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func economicsDenom(d model.Report) string {
	if d.EVMDenom != "" {
		return d.EVMDenom
	}
	if d.BondDenom != "" {
		return d.BondDenom
	}
	return "apmt"
}

func splitOutstandingSuffix(s string) (amount, suffix string) {
	if i := strings.Index(s, "  across "); i >= 0 {
		return strings.TrimSpace(s[:i]), strings.TrimSpace(s[i+2:])
	}
	return strings.TrimSpace(s), ""
}

func economicsPMTPerBlock(d model.Report) (perBlock float64, unit string, ok bool) {
	if !d.PMTEnabled || d.PMTRate == "" {
		return 0, "", false
	}
	s := strings.TrimSuffix(strings.TrimSpace(d.PMTRate), "/block")
	parts := strings.Fields(s)
	if len(parts) == 0 {
		return 0, "", false
	}
	v, err := strconv.ParseFloat(parts[0], 64)
	if err != nil || v <= 0 {
		return 0, "", false
	}
	unit = economicsDenom(d)
	if len(parts) >= 2 {
		unit = parts[1]
	}
	return v, unit, true
}

func moduleAccountBalance(d model.Report, name string) string {
	for _, m := range d.ModuleAccounts {
		if m.Name == name {
			return m.Balance
		}
	}
	return ""
}

func FeeCollectorBalance(d model.Report) string {
	return moduleAccountBalance(d, "fee_collector")
}

func distributionModuleBalance(d model.Report) string {
	return moduleAccountBalance(d, "distribution")
}

func economicsParseAmount(s string) (float64, bool) {
	s = strings.TrimSpace(s)
	if s == "" || s == "—" {
		return 0, false
	}
	if i := strings.Index(s, "  _"); i >= 0 {
		s = strings.TrimSpace(s[:i])
	}
	s = strings.TrimSuffix(s, "/block")
	parts := strings.Fields(strings.TrimPrefix(s, "~"))
	if len(parts) == 0 {
		return 0, false
	}
	v, err := strconv.ParseFloat(parts[0], 64)
	return v, err == nil
}

func rewardInPerBlockAmounts(d model.Report) (total float64, unit string, parts int) {
	unit = economicsDenom(d)
	if v, u, ok := economicsPMTPerBlock(d); ok {
		total += v
		unit = u
		parts++
	}
	if d.InflationPerBlock != "" {
		if v, ok := economicsParseAmount(d.InflationPerBlock); ok {
			total += v
			parts++
		}
	}
	if d.LastBlockFees != "" {
		if v, ok := economicsParseAmount(d.LastBlockFees); ok {
			total += v
			parts++
		}
	}
	return total, unit, parts
}

func RewardInPerBlockTotal(d model.Report) string {
	total, unit, parts := rewardInPerBlockAmounts(d)
	if parts == 0 || total <= 0 {
		return "—"
	}
	return fetch.FormatAmountUnit(total, unit) + "/block"
}

func economicsUnclaimedDelegator(d model.Report) string {
	if d.UnclaimedDelegator != "" {
		return d.UnclaimedDelegator
	}
	if d.Local.IsValidator && d.Local.Outstanding != "" {
		return d.Local.Outstanding
	}
	amt, _ := splitOutstandingSuffix(d.TotalOutstanding)
	return amt
}

func economicsUnclaimedCommission(d model.Report) string {
	if d.UnclaimedCommission != "" {
		return d.UnclaimedCommission
	}
	if d.Local.IsValidator && d.Local.CommissionEarned != "" {
		return d.Local.CommissionEarned
	}
	return ""
}

func economicsUnclaimedTotal(d model.Report) string {
	del := economicsUnclaimedDelegator(d)
	comm := economicsUnclaimedCommission(d)
	if del == "" && comm == "" {
		amt, _ := splitOutstandingSuffix(d.TotalOutstanding)
		return amt
	}
	if del != "" && comm != "" {
		delF, _ := economicsParseAmount(del)
		commF, _ := economicsParseAmount(comm)
		unit := economicsDenom(d)
		if p := strings.Fields(del); len(p) >= 2 {
			unit = p[1]
		}
		return fetch.FormatAmountUnit(delF+commF, unit)
	}
	if del != "" {
		return del
	}
	return comm
}

func economicsCommissionPct(d model.Report) (pct float64, ok bool) {
	if d.Local.IsValidator && d.Local.Commission > 0 {
		return d.Local.Commission, true
	}
	var sum float64
	var n int
	for _, v := range d.Validators {
		if v.CommissionFloat > 0 {
			sum += v.CommissionFloat
			n++
		}
	}
	if n > 0 {
		return sum / float64(n), true
	}
	return 0, false
}

// economicsPerBlockSplit estimates per-block community tax and per-validator op/delegator
// slices from PMTRate (Cosmos: tax first, then VP-weighted validator share, then commission).
func economicsPerBlockSplit(d model.Report) (tax, valPool, op, del float64, unit string, ok bool) {
	perBlock, unit, ok := economicsPMTPerBlock(d)
	if !ok {
		return 0, 0, 0, 0, "", false
	}
	tax = perBlock * d.CommunityTaxPct / 100
	valPool = perBlock - tax
	commPct, hasComm := economicsCommissionPct(d)
	if !hasComm {
		commPct = 0
	}
	n := float64(d.BondedCount)
	if n <= 0 && len(d.Validators) > 0 {
		n = float64(len(d.Validators))
	}
	if n <= 0 {
		n = 1
	}
	perVal := valPool / n
	op = perVal * commPct / 100
	del = perVal - op
	return tax, valPool, op, del, unit, true
}

func economicsFormatPerBlock(v float64, unit string) string {
	if v <= 0 || unit == "" {
		return "—"
	}
	return fetch.FormatAmountUnit(v, unit) + "/block"
}

func economicsFormatPerDay(d model.Report, perBlock float64, unit string) string {
	if perBlock <= 0 || unit == "" || d.BlockInterval == "" {
		return "—"
	}
	if d.PMTDailyEmit != "" && d.PMTRate != "" {
		perBlockPMT, u, ok := economicsPMTPerBlock(d)
		if ok && u == unit && perBlockPMT > 0 {
			ratio := perBlock / perBlockPMT
			if ratio > 0 && strings.HasPrefix(d.PMTDailyEmit, "~") {
				emit := strings.TrimPrefix(d.PMTDailyEmit, "~")
				emit = strings.TrimSuffix(emit, "/day")
				if f, err := strconv.ParseFloat(strings.Fields(emit)[0], 64); err == nil {
					return "~" + fetch.FormatAmountUnit(f*ratio, unit) + "/day"
				}
			}
		}
	}
	return "—"
}

func economicsLocalVPShare(d model.Report) string {
	if !d.Local.IsValidator || d.Local.VPPercent <= 0 {
		return ""
	}
	return fmt.Sprintf("%.2f%% VP", d.Local.VPPercent)
}

func economicsFeeCollectorCheck(d model.Report) string {
	bal := FeeCollectorBalance(d)
	if bal == "" {
		return "—"
	}
	v, ok := economicsParseAmount(bal)
	if !ok {
		return "—"
	}
	if v < 1e-6 {
		return "cleared"
	}
	return "not cleared?"
}

func economicsUnclaimedCheck(d model.Report) string {
	total := economicsUnclaimedTotal(d)
	if total == "" || d.TotalOutstanding == "" {
		return "—"
	}
	t1, ok1 := economicsParseAmount(total)
	t2, ok2 := economicsParseAmount(d.TotalOutstanding)
	if !ok1 || !ok2 {
		return "—"
	}
	if math.Abs(t1-t2) < 1e-6 {
		return "sums match"
	}
	return "—"
}

func economicsPMTPoolCheck(d model.Report) string {
	if !d.PMTEnabled {
		return "—"
	}
	if d.PMTPoolEmpty {
		return "pool empty"
	}
	return "OK"
}

func economicsLedgerRows(d model.Report) [][]string {
	var rows [][]string

	if d.PMTEnabled && d.PMTRate != "" {
		poolBal := d.PMTBalance
		if poolBal == "" {
			poolBal = "—"
		}
		rows = append(rows, []string{
			"1",
			"x/pmtrewards → fee_collector",
			d.PMTRate,
			poolBal,
			economicsPMTPoolCheck(d),
		})
	}

	if d.InflationPerBlock != "" {
		rows = append(rows, []string{
			"2",
			"x/mint inflation",
			d.InflationPerBlock,
			"—",
			fmt.Sprintf("%.2f%% inflation", d.Inflation),
		})
	} else if d.Inflation == 0 {
		rows = append(rows, []string{
			"2",
			"x/mint inflation",
			"0",
			"—",
			"inactive",
		})
	}

	if d.LastBlockFees != "" {
		rows = append(rows, []string{
			"3",
			"tx fees (parent block)",
			d.LastBlockFees,
			"—",
			"gas used × base fee",
		})
	} else if d.MempoolTxs > 0 {
		rows = append(rows, []string{
			"3",
			"tx fees",
			"—",
			"—",
			fmt.Sprintf("mempool %d pending", d.MempoolTxs),
		})
	}

	inBlock := RewardInPerBlockTotal(d)
	feeBal := FeeCollectorBalance(d)
	if feeBal == "" {
		feeBal = "—"
	}
	if inBlock != "—" || feeBal != "—" {
		rows = append(rows, []string{
			"4",
			"fee_collector",
			inBlock,
			feeBal,
			economicsFeeCollectorCheck(d),
		})
	}

	tax, valPool, op, del, unit, splitOK := economicsPerBlockSplit(d)
	if splitOK || d.CommunityTax != "" {
		taxStr, valStr := "—", "—"
		if splitOK {
			taxStr = economicsFormatPerBlock(tax, unit)
			valStr = economicsFormatPerBlock(valPool, unit)
		}
		rows = append(rows, []string{
			"5",
			"community tax → pool",
			taxStr,
			d.CommunityPool,
			d.CommunityTax + " of rewards",
		})
		distBal := distributionModuleBalance(d)
		if distBal == "" {
			distBal = "—"
		}
		rows = append(rows, []string{
			"6",
			"validator pool (all vals)",
			valStr,
			distBal,
			"escrow until paid out",
		})
		if pct, has := economicsCommissionPct(d); has {
			rows = append(rows, []string{
				"7a",
				"→ operator commission (network)",
				economicsFormatPerBlock(op, unit),
				economicsUnclaimedCommission(d),
				fmt.Sprintf("%.1f%% commission (avg)", pct),
			})
			rows = append(rows, []string{
				"7b",
				"→ delegator rewards (network)",
				economicsFormatPerBlock(del, unit),
				economicsUnclaimedDelegator(d),
				economicsUnclaimedCheck(d),
			})
		}
	} else if d.CommunityPool != "" {
		rows = append(rows, []string{
			"5",
			"community pool",
			"—",
			d.CommunityPool,
			d.CommunityTax,
		})
	}

	rows = appendEconomicsLocalLedgerRows(rows, d)
	return rows
}

func appendEconomicsLocalLedgerRows(rows [][]string, d model.Report) [][]string {
	if !d.Local.IsValidator {
		return rows
	}
	lv := d.Local
	_, valPool, _, _, unit, ok := economicsPerBlockSplit(d)
	if ok && lv.VPPercent > 0 {
		vp := lv.VPPercent / 100
		localShare := valPool * vp
		localOp := localShare * lv.Commission / 100
		localDel := localShare - localOp
		rows = append(rows, []string{
			"8a",
			"this validator → commission",
			economicsFormatPerBlock(localOp, unit),
			lv.CommissionEarned,
			fmt.Sprintf("%.2f%% VP · %.2f%% commission", lv.VPPercent, lv.Commission),
		})
		rows = append(rows, []string{
			"8b",
			"this validator → delegators",
			economicsFormatPerBlock(localDel, unit),
			lv.Outstanding,
			economicsLocalVPShare(d),
		})
		return rows
	}
	bal := lv.CommissionEarned
	if lv.Outstanding != "" {
		if bal != "" {
			bal += " / " + lv.Outstanding
		} else {
			bal = lv.Outstanding
		}
	}
	if bal == "" {
		bal = "—"
	}
	label := "this validator"
	if lv.Moniker != "" {
		label += " (" + lv.Moniker + ")"
	}
	rows = append(rows, []string{
		"8",
		label,
		"—",
		bal,
		economicsLocalVPShare(d),
	})
	return rows
}
