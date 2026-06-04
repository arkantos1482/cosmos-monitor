package markdown

import (
	"strings"
	"testing"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func TestBuildMermaidFence(t *testing.T) {
	d := model.Report{Inflation: 3.5, PMTEnabled: true, PMTRate: "0.1 PMT/block"}
	md := Build(d)
	if !strings.Contains(md, "```mermaid") {
		t.Fatal("markdown should use mermaid fenced blocks")
	}
	if strings.Contains(md, `<div class="mermaid">`) {
		t.Fatal("markdown should not embed raw mermaid HTML")
	}
	if !strings.Contains(md, "graph LR") {
		t.Fatal("economics diagram should use LR layout")
	}
}

func TestBuildFeeMath(t *testing.T) {
	d := model.Report{
		BlockHeight: "100", BaseFee: "1", BaseFeeRaw: "1000",
		BlockGas: "21000", ParentBlockGasWanted: 21000,
		BlockGasLimit: 100_000_000, Elasticity: 2,
		BaseFeeChangeDenominator: 8, MinGasMultiplier: "0.5",
		ParentBlockResultsOK: true,
	}
	md := Build(d)
	if !strings.Contains(md, "\n$$\n") {
		t.Fatal("fee math should use $$ display blocks")
	}
	if strings.Contains(md, `class="fee-math-tex"`) {
		t.Fatal("markdown should not use HTML katex hooks")
	}
}

func TestEconomicsOverviewMermaidSyntax(t *testing.T) {
	d := model.Report{
		Inflation:        3.5,
		BondedPct:        72.3,
		BlockHeight:      "482,160",
		CommunityTax:     "2.00%",
		PMTEnabled:       true,
		PMTRate:          "0.1000 PMT/block",
		TotalOutstanding: "0.006854 PMT  across 4 validators",
		Validators:       []model.Validator{{CommissionFloat: 10}},
	}
	src := economicsOverviewMermaid(d)
	if strings.Contains(src, "\nheight ") || strings.Contains(src, "\nmempool ") {
		t.Fatal("node labels must use <br/> not literal newlines")
	}
	if !strings.Contains(src, "<br/>") {
		t.Fatal("node labels should use <br/> line breaks")
	}
	if !strings.Contains(src, `|"block 482,160"|`) {
		t.Fatal("edge labels with commas must be quoted for mermaid.js")
	}
}

func TestBuildGoldenMinimal(t *testing.T) {
	d := model.Report{
		Moniker: "node1", Synced: true, BlockHeight: "1",
		BondDenom: "apmt", PMTEnabled: false,
	}
	md := Build(d)
	if !strings.Contains(md, "# 1. INFRASTRUCTURE") {
		t.Fatal("expected infrastructure section")
	}
	if !strings.Contains(md, "# 7. EVM JSON-RPC") {
		t.Fatal("expected EVM section")
	}
}
