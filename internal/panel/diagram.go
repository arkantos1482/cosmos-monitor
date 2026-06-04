package panel

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/arkantos1482/cosmos-monitor/internal/fetch"
	"github.com/arkantos1482/cosmos-monitor/internal/model"
	"github.com/arkantos1482/cosmos-monitor/internal/report"
)

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
