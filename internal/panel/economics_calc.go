package panel

import (
	"fmt"
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
		delF, _ := strconv.ParseFloat(strings.Fields(del)[0], 64)
		commF, _ := strconv.ParseFloat(strings.Fields(comm)[0], 64)
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
	// BlockInterval is like "5.2s" from report.FormatDur — parse loosely via PMTRate path:
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
	return fmt.Sprintf("%.2f%% of bonded stake", d.Local.VPPercent)
}
