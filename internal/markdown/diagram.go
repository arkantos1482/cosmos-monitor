package markdown

import (
	"fmt"
	"github.com/arkantos1482/cosmos-monitor/internal/model"
	"github.com/arkantos1482/cosmos-monitor/internal/report"
	"io"
	"strconv"
	"strings"

	"github.com/arkantos1482/cosmos-monitor/internal/fetch"
)

func writeMermaidFence(w io.Writer, src string) {
	fmt.Fprintf(w, "```mermaid\n%s\n```\n\n", strings.TrimSpace(src))
}

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
	lines := []string{"validator rewards", "per block allocation"}
	if d.TotalOutstanding != "" {
		amt, suffix := splitOutstandingSuffix(d.TotalOutstanding)
		lines = append(lines, "unclaimed "+amt)
		if suffix != "" {
			lines = append(lines, suffix)
		}
	}
	return stackLabelText(lines...)
}

func economicsCommLabel(d model.Report) string {
	comm := fmt.Sprintf("community tax %s", d.CommunityTax)
	if d.CommunityTaxZero {
		comm = "community tax 0%"
	}
	if d.CommunityPool != "" {
		return stackLabelText(comm, "pool "+d.CommunityPool)
	}
	return stackLabelText(comm)
}

func economicsCommissionPct(d model.Report) (pct float64, ok bool) {
	if d.Local.IsValidator && d.Local.Commission > 0 {
		return d.Local.Commission, true
	}
	if n := len(d.Validators); n > 0 {
		sum := 0.0
		for _, v := range d.Validators {
			sum += v.CommissionFloat
		}
		return sum / float64(n), true
	}
	return 0, false
}

func economicsOpLabel(d model.Report) string {
	if pct, ok := economicsCommissionPct(d); ok {
		return stackLabelText(fmt.Sprintf("operator %.1f%%", pct), "commission slice")
	}
	return stackLabelText("validator operator", "commission slice")
}

func economicsDelLabel(d model.Report) string {
	lines := []string{"delegators", "remainder share"}
	if pct, ok := economicsCommissionPct(d); ok && pct < 100 {
		lines = append(lines, fmt.Sprintf("(1 − %.1f%%)", pct))
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
	lines := []string{"x/distribution BeginBlock"}
	if d.CommunityTax != "" {
		lines = append(lines, "community tax "+d.CommunityTax)
	}
	if d.TotalOutstanding != "" {
		amt, suffix := splitOutstandingSuffix(d.TotalOutstanding)
		lines = append(lines, "unclaimed "+amt)
		if suffix != "" {
			lines = append(lines, suffix)
		}
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
	if d.CommunityTaxZero {
		return "tax 0%"
	}
	return "tax " + d.CommunityTax
}

func economicsDistValEdge(d model.Report) string {
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

func economicsSplitEdgeLabels(d model.Report) (comm, del string) {
	if pct, ok := economicsCommissionPct(d); ok {
		return fmt.Sprintf("commission %.0f%%", pct), fmt.Sprintf("remainder %.0f%%", 100-pct)
	}
	return "commission", "remainder"
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

func writeEconomicsDiagram(w io.Writer, d model.Report) {
	writeMermaidFence(w, economicsOverviewMermaid(d))
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

func writeFeemarketDiagram(w io.Writer, d model.Report) {
	writeMermaidFence(w, feemarketMechanicsMermaid(d))
}
