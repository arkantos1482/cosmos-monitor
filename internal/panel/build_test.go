package panel

import (
	"strings"
	"testing"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func TestBuildMermaidDiv(t *testing.T) {
	d := model.Report{Inflation: 3.5, PMTEnabled: true, PMTRate: "0.1 PMT/block"}
	out := Build(d)
	if !strings.Contains(out, `class="diagram-panel mermaid"`) {
		t.Fatal("panel should emit mermaid divs")
	}
	if strings.Contains(out, "```mermaid") {
		t.Fatal("panel should not use markdown mermaid fences")
	}
	if !strings.Contains(out, "graph LR") {
		t.Fatal("economics diagram should use LR layout")
	}
}

func TestBuildFeeMarketPanel(t *testing.T) {
	d := model.Report{
		BlockHeight: "100", BaseFee: "1", BaseFeeRaw: "1000",
		BlockGas: "21000", ParentBlockGasWanted: 21000,
		BlockGasLimit: 100_000_000, Elasticity: 2,
		BaseFeeChangeDenominator: 8, MinGasMultiplier: "0.5",
		ParentBlockResultsOK: true,
	}
	out := Build(d)
	if !strings.Contains(out, `class="fee-traffic"`) {
		t.Fatal("fee market section should render fee-traffic panel")
	}
	if strings.Contains(out, `Fee market (x/feemarket)`) {
		idx := strings.Index(out, `Fee market (x/feemarket)`)
		chunk := out[idx:]
		if end := strings.Index(chunk, `PMT Rewards`); end > 0 {
			chunk = chunk[:end]
		}
		if strings.Contains(chunk, `math-panel`) || strings.Contains(chunk, `class="diagram-panel mermaid"`) {
			t.Fatal("fee market subsection should not use KaTeX or feemarket mermaid")
		}
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
	if !strings.Contains(out, `class="dash-heading">1. INFRASTRUCTURE</h2>`) {
		t.Fatal("expected infrastructure section")
	}
	if !strings.Contains(out, `class="dash-heading">7. EVM JSON-RPC</h2>`) {
		t.Fatal("expected EVM section")
	}
}

func TestContentInventory(t *testing.T) {
	d := model.Report{
		Moniker: "node1", Synced: true, BlockHeight: "482,160",
		PMTEnabled: true, PMTRate: "0.1 PMT/block",
		EVMHTTPEndpoint: "http://localhost:8545", EVMChainID: 290290,
	}
	out := Build(d)
	for _, want := range []string{
		`class="dash-heading">1. INFRASTRUCTURE</h2>`,
		`class="dash-heading">2. NODE</h2>`,
		`class="dash-heading">3. VALIDATOR SET</h2>`,
		`class="dash-heading">4. THIS VALIDATOR</h2>`,
		`class="dash-heading">5. ECONOMICS</h2>`,
		`class="dash-heading">6. GOVERNANCE</h2>`,
		`class="dash-heading">7. EVM JSON-RPC</h2>`,
		`class="dash-subheading">For operators</h3>`,
		`class="dash-subheading">Probe health</h3>`,
		`class="stat-grid"`,
		`class="data-table"`,
		"Fee market (x/feemarket)",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("HTML inventory missing %q", want)
		}
	}
}
