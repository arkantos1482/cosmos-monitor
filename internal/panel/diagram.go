package panel

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/arkantos1482/cosmos-monitor/internal/fetch"
	"github.com/arkantos1482/cosmos-monitor/internal/model"
	"github.com/arkantos1482/cosmos-monitor/internal/report"
)

func mermaidLabel(s string) string {
	s = strings.ReplaceAll(s, `"`, `'`)
	s = strings.ReplaceAll(s, "\n", " ")
	return `"` + s + `"`
}

func stackLabelText(parts ...string) string {
	var lines []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			lines = append(lines, p)
		}
	}
	s := strings.Join(lines, "\n")
	s = strings.ReplaceAll(s, `"`, `'`)
	return s
}

func diagramDenom(d model.Report) string {
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

// economicsPMTPerBlock parses display PMTRate (e.g. "0.1000 PMT/block").
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
	unit = diagramDenom(d)
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
		unit := diagramDenom(d)
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

func economicsEdgeAmount(v float64, unit string) string {
	if v <= 0 || unit == "" {
		return ""
	}
	return "~" + fetch.FormatAmountUnit(v, unit) + "/blk"
}

func economicsFeesLabel(d model.Report) string {
	lines := []string{"Tx fees (ante / EVM)"}
	lines = append(lines, fmt.Sprintf("mempool %d · evm %d+%d", d.MempoolTxs, d.PendingTx, d.QueuedTx))
	if d.GasPrice != "" {
		lines = append(lines, "gas "+fetch.FormatFeeAmount(d.GasPrice, diagramDenom(d)))
	}
	return stackLabelText(lines...)
}

func economicsFCLabel(d model.Report) string {
	lines := []string{"fee_collector", "cleared BeginBlock"}
	if d.BlockHeight != "" {
		lines = append(lines, "height "+d.BlockHeight)
	}
	return stackLabelText(lines...)
}

func economicsStakeLabel(d model.Report) string {
	lines := []string{"x/staking"}
	if d.BondedCount > 0 {
		lines = append(lines, fmt.Sprintf("%d validators", d.BondedCount))
	}
	if d.BondedAmt != "" {
		lines = append(lines, d.BondedAmt+" bonded")
	} else if d.BondedPct > 0 {
		lines = append(lines, fmt.Sprintf("%.1f%% of supply", d.BondedPct))
	}
	lines = append(lines, "defines voting power")
	return stackLabelText(lines...)
}

func economicsValLabel(d model.Report) string {
	lines := []string{"validator rewards", "VP-weighted share"}
	if total := economicsUnclaimedTotal(d); total != "" {
		lines = append(lines, "unclaimed "+total)
	}
	if _, suffix := splitOutstandingSuffix(d.TotalOutstanding); suffix != "" {
		lines = append(lines, suffix)
	} else if d.BondedCount > 0 {
		lines = append(lines, fmt.Sprintf("%d validators", d.BondedCount))
	}
	if _, valPool, _, _, unit, ok := economicsPerBlockSplit(d); ok {
		if s := economicsEdgeAmount(valPool, unit); s != "" {
			lines = append(lines, s+" to validators")
		}
	}
	return stackLabelText(lines...)
}

func economicsCommLabel(d model.Report) string {
	comm := fmt.Sprintf("community tax %s", d.CommunityTax)
	if d.CommunityTaxZero {
		comm = "community tax 0%"
	}
	lines := []string{comm}
	if d.CommunityPool != "" && d.CommunityPool != "0" {
		lines = append(lines, "pool "+d.CommunityPool)
	}
	if tax, _, _, _, unit, ok := economicsPerBlockSplit(d); ok && tax > 0 {
		lines = append(lines, economicsEdgeAmount(tax, unit))
	}
	return stackLabelText(lines...)
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

func economicsOpLabel(d model.Report) string {
	lines := []string{"validator operator", "accumulated commission"}
	if pct, ok := economicsCommissionPct(d); ok {
		lines[0] = fmt.Sprintf("operator %.1f%%", pct)
	}
	if d.Local.IsValidator && d.Local.CommissionEarned != "" {
		lines = append(lines, "unclaimed "+d.Local.CommissionEarned)
	} else if amt := economicsUnclaimedCommission(d); amt != "" {
		lines = append(lines, "unclaimed "+amt)
	}
	if _, _, op, _, unit, ok := economicsPerBlockSplit(d); ok && op > 0 {
		lines = append(lines, economicsEdgeAmount(op, unit)+" / val (avg)")
	}
	return stackLabelText(lines...)
}

func economicsDelLabel(d model.Report) string {
	lines := []string{"delegators", "outstanding_rewards"}
	if pct, ok := economicsCommissionPct(d); ok && pct < 100 {
		lines = append(lines, fmt.Sprintf("%.1f%% of val share", 100-pct))
	}
	if d.Local.IsValidator && d.Local.Outstanding != "" {
		lines = append(lines, "unclaimed "+d.Local.Outstanding)
	} else if amt := economicsUnclaimedDelegator(d); amt != "" {
		lines = append(lines, "unclaimed "+amt)
	}
	if _, _, _, del, unit, ok := economicsPerBlockSplit(d); ok && del > 0 {
		lines = append(lines, economicsEdgeAmount(del, unit)+" / val (avg)")
	}
	return stackLabelText(lines...)
}

func economicsPMTPoolLabel(d model.Report) string {
	lines := []string{"PMT pool (x/pmtrewards)"}
	if d.PMTRate != "" {
		lines = append(lines, d.PMTRate)
	}
	if d.PMTBalance != "" {
		lines = append(lines, d.PMTBalance)
		if d.PMTRunway != "" {
			lines = append(lines, d.PMTRunway)
		}
	} else if d.PMTPoolEmpty {
		lines = append(lines, "— empty")
	}
	return stackLabelText(lines...)
}

func economicsDistLabel(d model.Report) string {
	lines := []string{"x/distribution BeginBlock", "AllocateRewards"}
	if perBlock, unit, ok := economicsPMTPerBlock(d); ok {
		lines = append(lines, fetch.FormatAmountUnit(perBlock, unit)+"/block in")
	} else if d.Inflation > 0 {
		lines = append(lines, fmt.Sprintf("inflation %.2f%%", d.Inflation))
	}
	if d.CommunityTax != "" {
		lines = append(lines, "tax "+d.CommunityTax)
	}
	return stackLabelText(lines...)
}

func economicsInflEdge(d model.Report) string {
	if d.Inflation > 0 {
		return fmt.Sprintf("mint +%.2f%%", d.Inflation)
	}
	return "mint off (0%)"
}

func economicsPMTEdge(d model.Report) string {
	if d.PMTRate != "" {
		return d.PMTRate
	}
	return "mint hook"
}

func economicsDistCommEdge(d model.Report) string {
	if tax, _, _, _, unit, ok := economicsPerBlockSplit(d); ok {
		if tax > 0 {
			if s := economicsEdgeAmount(tax, unit); s != "" {
				return s
			}
		}
		if d.CommunityTaxZero || d.CommunityTaxPct == 0 {
			return "tax 0%"
		}
	}
	if d.CommunityTaxZero || d.CommunityTax == "" {
		return "tax 0%"
	}
	return "tax " + d.CommunityTax
}

func economicsDistValEdge(d model.Report) string {
	if _, valPool, _, _, unit, ok := economicsPerBlockSplit(d); ok {
		if s := economicsEdgeAmount(valPool, unit); s != "" {
			return s
		}
	}
	pct := 100.0 - d.CommunityTaxPct
	if pct <= 0 {
		return "to validators"
	}
	return fmt.Sprintf("%.0f%% to validators", pct)
}

func economicsInflLabel(d model.Report) string {
	var lines []string
	lines = append(lines, "x/mint BeginBlock")
	if d.Inflation > 0 {
		lines = append(lines, fmt.Sprintf("inflation %.2f%%", d.Inflation))
		if d.AnnualProvisions != "" {
			lines = append(lines, d.AnnualProvisions+"/yr")
		}
	} else {
		lines = append(lines, "0% inflation (inactive)")
	}
	if d.GoalBonded > 0 {
		lines = append(lines, fmt.Sprintf("goal bonded %.0f%%", d.GoalBonded))
	}
	return stackLabelText(lines...)
}

func economicsSplitEdgeLabels(d model.Report) (op, del string) {
	if _, _, opAmt, delAmt, unit, ok := economicsPerBlockSplit(d); ok {
		if s := economicsEdgeAmount(opAmt, unit); s != "" {
			op = s
		}
		if s := economicsEdgeAmount(delAmt, unit); s != "" {
			del = s
		}
	}
	if op == "" || del == "" {
		if pct, ok := economicsCommissionPct(d); ok {
			if op == "" {
				op = fmt.Sprintf("commission %.0f%%", pct)
			}
			if del == "" {
				del = fmt.Sprintf("delegators %.0f%%", 100-pct)
			}
		} else {
			if op == "" {
				op = "commission"
			}
			if del == "" {
				del = "delegators"
			}
		}
	}
	return op, del
}

func stackMermaidQuoted(label string) string {
	label = strings.ReplaceAll(label, `"`, `'`)
	label = strings.ReplaceAll(label, "\n", "<br/>")
	return `"` + label + `"`
}

func mermaidEdgeText(s string) string {
	if s == "" {
		return s
	}
	s = strings.ReplaceAll(s, `"`, `'`)
	return `"` + s + `"`
}

func writeStackNode(b *strings.Builder, id, label string) {
	fmt.Fprintf(b, "  %s[%s]\n", id, stackMermaidQuoted(label))
}

func writeEconomicsNodes(b *strings.Builder, d model.Report) {
	writeStackNode(b, "fees", economicsFeesLabel(d))
	writeStackNode(b, "infl", economicsInflLabel(d))
	if d.PMTEnabled {
		writeStackNode(b, "pmtPool", economicsPMTPoolLabel(d))
	}
	writeStackNode(b, "fc", economicsFCLabel(d))
	writeStackNode(b, "staking", economicsStakeLabel(d))
	writeStackNode(b, "dist", economicsDistLabel(d))
	writeStackNode(b, "comm", economicsCommLabel(d))
	writeStackNode(b, "val", economicsValLabel(d))
	writeStackNode(b, "op", economicsOpLabel(d))
	writeStackNode(b, "del", economicsDelLabel(d))
}

func writeEconomicsEdges(b *strings.Builder, d model.Report) {
	fmt.Fprintf(b, "  fees --> fc\n")
	fmt.Fprintf(b, "  infl -->|%s| fc\n", mermaidEdgeText(economicsInflEdge(d)))
	if d.PMTEnabled {
		fmt.Fprintf(b, "  pmtPool -->|%s| fc\n", mermaidEdgeText(economicsPMTEdge(d)))
	}
	if d.BlockHeight != "" {
		fmt.Fprintf(b, "  fc -->|%s| dist\n", mermaidEdgeText("block "+d.BlockHeight))
	} else {
		fmt.Fprintf(b, "  fc --> dist\n")
	}
	fmt.Fprintf(b, "  staking -->|%s| dist\n", mermaidEdgeText(fmt.Sprintf("%.1f%% bonded", d.BondedPct)))
	fmt.Fprintf(b, "  dist -->|%s| comm\n", mermaidEdgeText(economicsDistCommEdge(d)))
	fmt.Fprintf(b, "  dist -->|%s| val\n", mermaidEdgeText(economicsDistValEdge(d)))
	commEdge, delEdge := economicsSplitEdgeLabels(d)
	fmt.Fprintf(b, "  val -->|%s| op\n", mermaidEdgeText(commEdge))
	fmt.Fprintf(b, "  val -->|%s| del\n", mermaidEdgeText(delEdge))
}

func economicsOverviewMermaid(d model.Report) string {
	var b strings.Builder
	fmt.Fprintf(&b, "graph LR\n")
	writeEconomicsNodes(&b, d)
	fmt.Fprintf(&b, "  subgraph sources\n    fees\n    infl\n")
	if d.PMTEnabled {
		fmt.Fprintf(&b, "    pmtPool\n")
	}
	fmt.Fprintf(&b, "  end\n")
	writeEconomicsEdges(&b, d)
	return b.String()
}

func writeEconomicsDiagram(w Writer, d model.Report) {
	w.Mermaid(economicsOverviewMermaid(d))
}

func parseDiagramUint(s string) uint64 {
	s = strings.ReplaceAll(s, ",", "")
	n, _ := strconv.ParseUint(s, 10, 64)
	return n
}

func feemarketGasNumbers(d model.Report) (wanted, target uint64, ok bool) {
	wanted = feemarketStoredWanted(d)
	if d.BlockGasLimit > 0 && d.Elasticity > 0 {
		return wanted, d.BlockGasLimit / uint64(d.Elasticity), true
	}
	return wanted, 0, false
}

func feemarketCalcEdge(d model.Report) string {
	wanted, target, ok := feemarketGasNumbers(d)
	if !ok {
		return "vs target"
	}
	if wanted > target {
		return "↑"
	}
	if wanted < target {
		return "↓"
	}
	return "="
}

func feemarketSlimEndBlockLabel(d model.Report) string {
	lines := []string{"EndBlock N-1", "max(wanted×μ, gasUsed)"}
	if d.ParentBlockGasUsed > 0 {
		lines = append(lines, "used "+report.FormatInt(int64(d.ParentBlockGasUsed)))
	}
	return stackLabelText(lines...)
}

func feemarketSlimStoredLabel(d model.Report) string {
	wanted := feemarketStoredWanted(d)
	lines := []string{"stored W", "GetBlockGasWanted"}
	if wanted > 0 {
		lines = append(lines, report.FormatInt(int64(wanted)))
	}
	return stackLabelText(lines...)
}

func feemarketSlimCalcLabel(d model.Report) string {
	return stackLabelText("BeginBlock N", "CalculateBaseFee")
}

func feemarketSlimBaseFeeLabel(d model.Report) string {
	if d.NoBaseFee {
		return stackLabelText("no_base_fee")
	}
	lines := []string{"base fee @ N"}
	if d.BaseFee != "" {
		lines = append(lines, d.BaseFee)
	}
	return stackLabelText(lines...)
}

func feemarketSlimAnteLabel(d model.Report) string {
	denom := diagramDenom(d)
	lines := []string{"ante · eth_gasPrice"}
	if d.GasPrice != "" {
		lines = append(lines, fetch.FormatFeeAmount(d.GasPrice, denom))
	}
	lines = append(lines, fmt.Sprintf("mempool %d", d.MempoolTxs))
	return stackLabelText(lines...)
}

func writeFeemarketSlimNodes(b *strings.Builder, d model.Report) {
	writeStackNode(b, "endBlk", feemarketSlimEndBlockLabel(d))
	writeStackNode(b, "stored", feemarketSlimStoredLabel(d))
	writeStackNode(b, "calc", feemarketSlimCalcLabel(d))
	writeStackNode(b, "baseFee", feemarketSlimBaseFeeLabel(d))
	writeStackNode(b, "ante", feemarketSlimAnteLabel(d))
}

func writeFeemarketSlimEdges(b *strings.Builder, d model.Report) {
	fmt.Fprintf(b, "  endBlk -->|%s| stored\n", mermaidEdgeText("prior block"))
	fmt.Fprintf(b, "  stored --> calc\n")
	fmt.Fprintf(b, "  calc -->|%s| baseFee\n", mermaidEdgeText(feemarketCalcEdge(d)))
	fmt.Fprintf(b, "  baseFee --> ante\n")
	fmt.Fprintf(b, "  ante --> endBlk\n")
}

func feemarketMechanicsMermaid(d model.Report) string {
	var b strings.Builder
	fmt.Fprintf(&b, "graph LR\n")
	writeFeemarketSlimNodes(&b, d)
	fmt.Fprintf(&b, "  subgraph endBlock[\"EndBlock N-1\"]\n    endBlk\n    stored\n  end\n")
	fmt.Fprintf(&b, "  subgraph beginBlock[\"BeginBlock N\"]\n    calc\n    baseFee\n  end\n")
	fmt.Fprintf(&b, "  subgraph execution[\"Block N\"]\n    ante\n  end\n")
	writeFeemarketSlimEdges(&b, d)
	return b.String()
}

func writeFeemarketDiagram(w Writer, d model.Report) {
	w.Mermaid(feemarketMechanicsMermaid(d))
}
