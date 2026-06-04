package panel

import (
	"strings"
	"testing"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func TestBuildEconomicsUsesTablesNotMermaid(t *testing.T) {
	d := model.Report{
		Inflation: 3.5, PMTEnabled: true, PMTRate: "0.1 PMT/block",
		BaseFee: "1000", Elasticity: 2, BlockGas: "21000",
		ParentBlockGasWanted: 21000, ParentBlockResultsOK: true,
		ModuleAccounts: []model.ModuleAccountRow{{Name: "fee_collector", Balance: "1 PMT"}},
	}
	out := Build(d)
	idx := strings.Index(out, "5. ECONOMICS")
	end := strings.Index(out, "6. GOVERNANCE")
	if idx < 0 || end < 0 {
		t.Fatal("expected economics and governance sections")
	}
	eco := out[idx:end]
	if strings.Contains(eco, `class="diagram-panel mermaid"`) || strings.Contains(eco, "graph LR") {
		t.Fatal("economics section should not use mermaid")
	}
	if !strings.Contains(eco, "Money flow (live balances)") {
		t.Fatal("economics should use live balance tables")
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
		if strings.Contains(chunk, `math-panel`) {
			t.Fatal("fee market subsection should not use KaTeX")
		}
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
		ModuleAccounts: []model.ModuleAccountRow{{Name: "fee_collector", Balance: "1 PMT"}},
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
		"Money flow (live balances)",
		"Fee market (x/feemarket)",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("HTML inventory missing %q", want)
		}
	}
}
