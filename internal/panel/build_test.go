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
	end := strings.Index(out, "6. FEE MARKET")
	if idx < 0 || end < 0 {
		t.Fatal("expected economics and governance sections")
	}
	eco := out[idx:end]
	if strings.Contains(eco, `class="diagram-panel mermaid"`) || strings.Contains(eco, "graph LR") {
		t.Fatal("economics section should not use mermaid")
	}
	if !strings.Contains(eco, "Block reward ledger") {
		t.Fatal("economics should use block reward ledger")
	}
	if !strings.Contains(eco, "At a glance") {
		t.Fatal("economics should include at-a-glance stats")
	}
	if !strings.Contains(eco, `economics-kpi-band`) {
		t.Fatal("economics should wrap at-a-glance in KPI band")
	}
	if !strings.Contains(eco, `data-table--ledger`) {
		t.Fatal("ledger table should use ledger styling")
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
	if !strings.Contains(out, `class="fee-hero"`) {
		t.Fatal("fee market section should render fee-hero panel")
	}
	if !strings.Contains(out, `class="fee-traffic"`) {
		t.Fatal("fee market section should render fee-traffic panel")
	}
	if !strings.Contains(out, `fee-key-metrics`) {
		t.Fatal("fee market hero should include KPI row")
	}
	idx := strings.Index(out, "6. FEE MARKET")
	end := strings.Index(out, "7. GOVERNANCE")
	if idx < 0 || end < 0 {
		t.Fatal("expected fee market and governance sections")
	}
	fee := out[idx:end]
	if strings.Contains(fee, `math-panel`) {
		t.Fatal("fee market section should not use KaTeX")
	}
	if !strings.Contains(out, `dash-section--feemarket`) {
		t.Fatal("fee market section should have feemarket accent class")
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
	if !strings.Contains(out, `dash-section--infra`) {
		t.Fatal("expected infrastructure section accent")
	}
	if !strings.Contains(out, `class="dash-heading">8. EVM JSON-RPC</h2>`) {
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
		`class="dash-heading">6. FEE MARKET</h2>`,
		`class="dash-heading">7. GOVERNANCE</h2>`,
		`class="dash-heading">8. EVM JSON-RPC</h2>`,
		`class="dash-subheading">For operators</h3>`,
		`class="dash-subheading">Probe health</h3>`,
		`class="kpi-grid"`,
		`class="kpi-tile"`,
		`class="data-table"`,
		"At a glance",
		"Block reward ledger",
		`class="dash-subheading">Chain parameters (reference)</h3>`,
		`dash-section--feemarket`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("HTML inventory missing %q", want)
		}
	}
}
