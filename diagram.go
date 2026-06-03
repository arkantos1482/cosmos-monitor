package main

import (
	"fmt"
	"io"
	"strings"

	mcmd "github.com/AlexanderGrooff/mermaid-ascii/cmd"
	"github.com/AlexanderGrooff/mermaid-ascii/pkg/diagram"
	"github.com/arkantos1482/cosmos-monitor/fetch"
)

// Diagram padding (mermaid-ascii): border = space inside each box; padX/padY = gap between nodes.
// Defaults are compact for TUI scrolling; override with -diagram-border / -diagram-padx / -diagram-pady.
var (
	diagramBorderPad = 0
	diagramPadX      = 2
	diagramPadY      = 2
)

func SetDiagramPadding(border, padX, padY int) {
	if border < 0 {
		border = 0
	}
	if padX < 0 {
		padX = 0
	}
	if padY < 0 {
		padY = 0
	}
	diagramBorderPad = border
	diagramPadX = padX
	diagramPadY = padY
}

func mermaidConfig(useAscii bool) (*diagram.Config, error) {
	return diagram.NewCLIConfig(useAscii, false, false, diagramBorderPad, diagramPadX, diagramPadY, "TD")
}

func mermaidConfigDirection(useAscii bool, direction string, padX, padY int) (*diagram.Config, error) {
	return diagram.NewCLIConfig(useAscii, false, false, diagramBorderPad, padX, padY, direction)
}

// renderMermaid converts Mermaid source to Unicode box-drawing text (terminal TUI).
func renderMermaid(src string) (string, error) {
	return renderMermaidWith(src, diagramPadX, diagramPadY, "TD")
}

func renderMermaidWith(src string, padX, padY int, direction string) (string, error) {
	cfg, err := mermaidConfigDirection(false, direction, padX, padY)
	if err != nil {
		return "", err
	}
	out, err := mcmd.RenderDiagram(src, cfg)
	if err == nil {
		return strings.TrimRight(out, "\n"), nil
	}
	cfg2, err2 := mermaidConfigDirection(true, direction, padX, padY)
	if err2 != nil {
		return "", err
	}
	out, err = mcmd.RenderDiagram(src, cfg2)
	return strings.TrimRight(out, "\n"), err
}

func writeMermaidWeb(w io.Writer, src string) {
	fmt.Fprintf(w, "<div class=\"mermaid\">\n%s</div>\n\n", src)
}

func writeDiagram(w io.Writer, mermaid string, web bool) {
	if web {
		writeMermaidWeb(w, mermaid)
		return
	}
	out, err := renderMermaid(mermaid)
	if err != nil {
		fmt.Fprintf(w, "_diagram render failed: %v_\n\n", err)
		return
	}
	fmt.Fprintf(w, "```text\n%s\n```\n\n", out)
}

func writeEconomicsDiagram(w io.Writer, d WebData, web bool) {
	src := economicsOverviewMermaid(d)
	if web {
		writeMermaidWeb(w, src)
		return
	}
	out, err := renderMermaid(economicsOverviewMermaidASCII(d))
	if err != nil {
		fmt.Fprintf(w, "_diagram render failed: %v_\n\n", err)
		return
	}
	fmt.Fprintf(w, "```text\n%s\n```\n\n", out)
}

func mermaidLabel(s string) string {
	s = strings.ReplaceAll(s, `"`, `'`)
	s = strings.ReplaceAll(s, "\n", " ")
	return `"` + s + `"`
}

// stackLabelText joins parts with newlines for taller, narrower mermaid-ascii boxes.
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

func diagramDenom(d WebData) string {
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

func economicsFeesLabel(d WebData) string {
	lines := []string{"Tx fees (ante / EVM)"}
	if d.MempoolTxs > 0 {
		lines = append(lines, fmt.Sprintf("mempool %d", d.MempoolTxs))
	}
	if d.PendingTx > 0 {
		lines = append(lines, fmt.Sprintf("evm pending %d", d.PendingTx))
	}
	return stackLabelText(lines...)
}

func economicsFCLabel(d WebData) string {
	return stackLabelText("fee_collector", "cleared each BeginBlock")
}

func economicsStakeLabel(d WebData) string {
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

func economicsValLabel(d WebData) string {
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

func economicsCommLabel(d WebData) string {
	comm := fmt.Sprintf("community tax %s", d.CommunityTax)
	if d.CommunityTaxZero {
		comm = "community tax 0%"
	}
	if d.CommunityPool != "" {
		return stackLabelText(comm, "pool "+d.CommunityPool)
	}
	return stackLabelText(comm)
}

func economicsCommissionPct(d WebData) (pct float64, ok bool) {
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

func economicsOpLabel(d WebData) string {
	if pct, ok := economicsCommissionPct(d); ok {
		return stackLabelText(fmt.Sprintf("operator %.1f%%", pct), "commission slice")
	}
	return stackLabelText("validator operator", "commission slice")
}

func economicsDelLabel(d WebData) string {
	lines := []string{"delegators", "remainder share"}
	if pct, ok := economicsCommissionPct(d); ok && pct < 100 {
		lines = append(lines, fmt.Sprintf("(1 − %.1f%%)", pct))
	}
	return stackLabelText(lines...)
}

func economicsPMTPoolLabel(d WebData) string {
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

func economicsDistLabel(d WebData) string {
	return stackLabelText(
		"x/distribution BeginBlock",
		"allocates block rewards",
		"reads x/staking",
	)
}

func economicsInflLabel(d WebData) string {
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

func economicsSplitEdgeLabels(d WebData) (comm, del string) {
	if pct, ok := economicsCommissionPct(d); ok {
		return fmt.Sprintf("commission %.0f%%", pct), fmt.Sprintf("remainder %.0f%%", 100-pct)
	}
	return "commission", "remainder"
}

func stackMermaidQuoted(label string) string {
	label = strings.ReplaceAll(label, `"`, `'`)
	return `"` + label + `"`
}

func writeStackNode(b *strings.Builder, id, label string) {
	fmt.Fprintf(b, "  %s[%s]\n", id, stackMermaidQuoted(label))
}

func writeEconomicsNodes(b *strings.Builder, d WebData) {
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

func writeEconomicsEdges(b *strings.Builder, d WebData) {
	fmt.Fprintf(b, "  fees --> fc\n")
	fmt.Fprintf(b, "  infl -->|mint if active| fc\n")
	if d.PMTEnabled {
		fmt.Fprintf(b, "  pmtPool -->|mint hook| fc\n")
	}
	fmt.Fprintf(b, "  fc --> dist\n")
	fmt.Fprintf(b, "  staking -->|voting power| dist\n")
	fmt.Fprintf(b, "  dist --> comm\n")
	fmt.Fprintf(b, "  dist --> val\n")
	commEdge, delEdge := economicsSplitEdgeLabels(d)
	fmt.Fprintf(b, "  val -->|%s| op\n", commEdge)
	fmt.Fprintf(b, "  val -->|%s| del\n", delEdge)
}

// economicsOverviewMermaid is the full LR graph for mermaid.js (web).
func economicsOverviewMermaid(d WebData) string {
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

// economicsOverviewMermaidASCII is a TD spine without subgraphs — mermaid-ascii friendly (terminal).
func economicsOverviewMermaidASCII(d WebData) string {
	var b strings.Builder
	fmt.Fprintf(&b, "graph TD\n")
	writeEconomicsNodes(&b, d)
	writeEconomicsEdges(&b, d)
	return b.String()
}

func feemarketMechanicsMermaid(d WebData) string {
	denom := diagramDenom(d)

	gasUsed := "gas used (last block)"
	if d.BlockGas != "" {
		gasUsed = "gas used: " + d.BlockGas
	}
	gasTarget := "target = max_block_gas ÷ elasticity"
	if d.Elasticity > 0 {
		gasTarget = fmt.Sprintf("target = max_block_gas ÷ %d", d.Elasticity)
	}

	parentBF := "parent base fee"
	if d.BaseFee != "" {
		parentBF = "parent base fee: " + fetch.FormatFeeAmount(d.BaseFee, denom)
	}

	baseFeeLabel := "base fee (this block)"
	if d.NoBaseFee {
		baseFeeLabel = "base fee disabled (no_base_fee)"
	} else if d.BaseFee != "" {
		baseFeeLabel = "base fee: " + fetch.FormatFeeAmount(d.BaseFee, denom)
	}

	gasRPC := "eth_gasPrice"
	if d.GasPrice != "" {
		gasRPC = "gas price: " + fetch.FormatFeeAmount(d.GasPrice, denom)
	}

	var paramsTune []string
	if d.BaseFeeChangeDenominator > 0 {
		paramsTune = append(paramsTune, fmt.Sprintf("change denom %d", d.BaseFeeChangeDenominator))
	}
	if d.Elasticity > 0 {
		paramsTune = append(paramsTune, fmt.Sprintf("elasticity %d", d.Elasticity))
	}
	paramsTuneLabel := "feemarket tuning"
	if len(paramsTune) > 0 {
		paramsTuneLabel = strings.Join(paramsTune, " · ")
	}

	var paramsFloor []string
	if d.MinGasPrice != "" {
		paramsFloor = append(paramsFloor, "min_gas "+d.MinGasPrice)
	}
	if d.AdjCap != "" {
		paramsFloor = append(paramsFloor, "max Δ "+d.AdjCap)
	}
	paramsFloorLabel := "feemarket floors"
	if len(paramsFloor) > 0 {
		paramsFloorLabel = strings.Join(paramsFloor, " · ")
	}

	var b strings.Builder
	fmt.Fprintf(&b, "graph TD\n")
	fmt.Fprintf(&b, "  gasUsed[%s]\n", mermaidLabel(gasUsed))
	fmt.Fprintf(&b, "  gasTarget[%s]\n", mermaidLabel(gasTarget))
	fmt.Fprintf(&b, "  parentBF[%s]\n", mermaidLabel(parentBF))
	fmt.Fprintf(&b, "  paramsTune[%s]\n", mermaidLabel(paramsTuneLabel))
	fmt.Fprintf(&b, "  paramsFloor[%s]\n", mermaidLabel(paramsFloorLabel))
	fmt.Fprintf(&b, "  calc[%s]\n", mermaidLabel("BeginBlock: CalculateBaseFee"))
	fmt.Fprintf(&b, "  baseFee[%s]\n", mermaidLabel(baseFeeLabel))
	fmt.Fprintf(&b, "  eff[%s]\n", mermaidLabel("effective price ≥ base fee"))
	fmt.Fprintf(&b, "  ante[%s]\n", mermaidLabel("ante: VerifyFee + DeductFees"))
	fmt.Fprintf(&b, "  gasRPC[%s]\n", mermaidLabel(gasRPC))
	fmt.Fprintf(&b, "  endBlk[%s]\n", mermaidLabel("EndBlock: block_gas_wanted"))

	// Vertical spine (avoids wide fan-in to calc / baseFee).
	fmt.Fprintf(&b, "  gasUsed --> gasTarget\n")
	fmt.Fprintf(&b, "  gasTarget --> parentBF\n")
	fmt.Fprintf(&b, "  parentBF --> paramsTune\n")
	fmt.Fprintf(&b, "  paramsTune --> paramsFloor\n")
	fmt.Fprintf(&b, "  paramsFloor --> calc\n")
	fmt.Fprintf(&b, "  calc -->|gasUsed > target ⇒ fee ↑| baseFee\n")
	fmt.Fprintf(&b, "  baseFee --> eff\n")
	fmt.Fprintf(&b, "  eff --> ante\n")
	fmt.Fprintf(&b, "  baseFee --> gasRPC\n")
	fmt.Fprintf(&b, "  ante --> endBlk\n")
	return b.String()
}
