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
		ModuleAccounts: []model.ModuleAccountRow{
			{Name: "fee_collector", Balance: "1 PMT", Role: "fees"},
			{Name: "distribution", Balance: "0 PMT", Role: "escrow"},
		},
		CommunityTax: "2%",
		CommunityPool: "0.5 PMT",
	}
	out := Build(d)
	idx := strings.Index(out, "3. ECONOMICS")
	end := strings.Index(out, `class="dash-heading">4. FEE MARKET</h2>`)
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
	if strings.Contains(eco, "At a glance") {
		t.Fatal("economics should not duplicate at-a-glance subsection")
	}
	if !strings.Contains(out, `eco-domains`) {
		t.Fatal("economics summary should use domain cards")
	}
	if !strings.Contains(out, `class="eco-summary"`) {
		t.Fatal("economics should include embedded summary")
	}
	if !strings.Contains(eco, `data-table--ledger`) {
		t.Fatal("ledger table should use ledger styling")
	}
	
	for _, want := range []string{
		"eco-domain--pmtrewards",
		"eco-domain--inflation",
		`class="dash-subheading">Distribution</h3>`,
	} {
		if !strings.Contains(eco, want) {
			t.Fatalf("economics should include %q", want)
		}
	}
	for _, gone := range []string{
		"eco-domain--staking",
		"eco-domain--slashing",
	} {
		if strings.Contains(eco, gone) {
			t.Fatalf("economics should not include %q", gone)
		}
	}
	if strings.Contains(eco, "Module accounts") || strings.Contains(eco, "eco-module-accounts") {
		t.Fatal("economics should not include unified module accounts table")
	}
	for _, gone := range []string{
		"eco-domain--txfees",
		"eco-domain--distribution",
		"eco-domain--rewards",
		`class="dash-subheading">Advanced parameters (reward flow)</h3>`,
		`class="dash-subheading">Chain parameters (reference)</h3>`,
	} {
		if strings.Contains(eco, gone) {
			t.Fatalf("economics should not contain %q", gone)
		}
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
	if !strings.Contains(out, `id="fee-L1"`) {
		t.Fatal("fee market section should render L1 ladder panel")
	}
	if strings.Contains(out, `class="fee-flow"`) {
		t.Fatal("fee market section should not use legacy fee-flow")
	}
	idx := strings.Index(out, `class="dash-heading">4. FEE MARKET</h2>`)
	end := strings.Index(out, "5. GOVERNANCE")
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
	if !strings.Contains(out, `class="dash-heading">3. EVM JSON-RPC</h2>`) {
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
		`class="dash-heading">1. STAKING</h2>`,
		`class="dash-heading">2. VALIDATOR SET</h2>`,
		`class="dash-heading">3. ECONOMICS</h2>`,
		`class="dash-heading">4. FEE MARKET</h2>`,
		`class="dash-heading">5. GOVERNANCE</h2>`,
		`class="dash-heading">1. INFRASTRUCTURE</h2>`,
		`class="dash-heading">2. VALIDATOR</h2>`,
		`class="dash-heading">3. EVM JSON-RPC</h2>`,
		`class="dash-subheading">For operators</h3>`,
		`class="dash-subheading">Probe health</h3>`,
		`class="kpi-grid"`,
		`class="kpi-tile"`,
		`class="data-table"`,
		`val-summary--p2p`,
		"Block reward ledger",
		`class="dash-subheading">Distribution</h3>`,
		"eco-domain--pmtrewards",
		`dash-section--feemarket`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("HTML inventory missing %q", want)
		}
	}
}
