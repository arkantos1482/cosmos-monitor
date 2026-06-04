package panel

import (
	"strings"
	"testing"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func TestBuildMermaidDiv(t *testing.T) {
	d := model.Report{Inflation: 3.5, PMTEnabled: true, PMTRate: "0.1 PMT/block"}
	out := Build(d)
	if !strings.Contains(out, `<div class="mermaid">`) {
		t.Fatal("panel should emit mermaid divs")
	}
	if strings.Contains(out, "```mermaid") {
		t.Fatal("panel should not use markdown mermaid fences")
	}
	if !strings.Contains(out, "graph LR") {
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
	out := Build(d)
	if !strings.Contains(out, `class="math-display"`) {
		t.Fatal("fee math should use math-display divs for KaTeX")
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
	out := Build(d)
	if !strings.Contains(out, "<h1>1. INFRASTRUCTURE</h1>") {
		t.Fatal("expected infrastructure section")
	}
	if !strings.Contains(out, "<h1>7. EVM JSON-RPC</h1>") {
		t.Fatal("expected EVM section")
	}
}

func TestBuildPlainPreservesSections(t *testing.T) {
	d := model.Report{Moniker: "n", Synced: true, BlockHeight: "1", BondDenom: "apmt"}
	plain := BuildText(d)
	if !strings.Contains(plain, "# 1. INFRASTRUCTURE") {
		t.Fatal("plain output should include infrastructure section")
	}
	if !strings.Contains(plain, "# 7. EVM JSON-RPC") {
		t.Fatal("plain output should include EVM section")
	}
}

func TestContentInventoryVsMarkdown(t *testing.T) {
	// Golden strings that must survive the HTML migration (sample from former markdown output).
	d := model.Report{
		Moniker: "node1", Synced: true, BlockHeight: "482,160",
		PMTEnabled: true, PMTRate: "0.1 PMT/block",
		EVMHTTPEndpoint: "http://localhost:8545", EVMChainID: 290290,
	}
	out := BuildText(d)
	for _, want := range []string{
		"1. INFRASTRUCTURE",
		"2. NODE",
		"3. VALIDATOR SET",
		"4. THIS VALIDATOR",
		"5. ECONOMICS",
		"6. GOVERNANCE",
		"7. EVM JSON-RPC",
		"For operators",
		"Probe health",
		"Fee market (x/feemarket)",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("plain inventory missing %q", want)
		}
	}
}
