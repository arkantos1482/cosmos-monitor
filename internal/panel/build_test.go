package panel

import (
	"strings"
	"testing"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func TestBuildRewardsUsesTablesNotMermaid(t *testing.T) {
	d := model.Report{
		Inflation: 3.5, PMTEnabled: true, PMTRate: "0.1 PMT/block",
		BaseFee: "1000", Elasticity: 2, BlockGas: "21000",
		ParentBlockGasWanted: 21000, ParentBlockResultsOK: true,
		ModuleAccounts: []model.ModuleAccountRow{
			{Name: "fee_collector", Balance: "1 PMT", Role: "fees"},
			{Name: "distribution", Balance: "0 PMT", Role: "escrow"},
		},
		CommunityTax:  "2%",
		CommunityPool: "0.5 PMT",
	}
	out := Build(d)
	rewardsIdx := strings.Index(out, "5. REWARDS")
	end := strings.Index(out, `class="dash-heading">6. GOVERNANCE</h2>`)
	if rewardsIdx < 0 || end < 0 {
		t.Fatal("expected rewards and governance sections")
	}
	rewards := out[rewardsIdx:end]
	if strings.Contains(rewards, `class="diagram-panel mermaid"`) || strings.Contains(rewards, "graph LR") {
		t.Fatal("rewards section should not use mermaid")
	}
	if strings.Contains(rewards, "Block reward ledger") {
		t.Fatal("rewards should not include block reward ledger")
	}
	if strings.Contains(rewards, "At a glance") {
		t.Fatal("rewards should not duplicate at-a-glance subsection")
	}
	if !strings.Contains(rewards, `eco-domains`) {
		t.Fatal("rewards summary should use domain cards")
	}
	if !strings.Contains(rewards, `class="eco-summary"`) {
		t.Fatal("rewards should include embedded summary")
	}

	for _, want := range []string{
		`class="dash-subheading">PMT Rewards</h3>`,
		`class="dash-subheading">Inflation</h3>`,
		"eco-domain--pmtrewards",
		"eco-domain--inflation",
	} {
		if !strings.Contains(rewards, want) {
			t.Fatalf("rewards should include %q", want)
		}
	}
	for _, gone := range []string{
		"eco-domain--staking",
		"eco-domain--slashing",
		"Module accounts",
		"eco-module-accounts",
		"eco-domain--txfees",
		"eco-domain--distribution",
		"eco-domain--rewards",
		`class="dash-subheading">Advanced parameters (reward flow)</h3>`,
		`class="dash-subheading">Chain parameters (reference)</h3>`,
		`class="dash-subheading">Routing</h3>`,
		`data-table--ledger`,
	} {
		if strings.Contains(rewards, gone) {
			t.Fatalf("rewards should not contain %q", gone)
		}
	}

	distIdx := strings.Index(out, "3. DISTRIBUTION")
	feeIdx := strings.Index(out, `class="dash-heading">4. FEE MARKET</h2>`)
	if distIdx < 0 || feeIdx < 0 {
		t.Fatal("expected distribution and fee market sections")
	}
	dist := out[distIdx:feeIdx]
	if !strings.Contains(dist, `data-table--ledger`) {
		t.Fatal("ledger table should live in distribution section")
	}
	if !strings.Contains(dist, `class="dash-subheading">Routing</h3>`) {
		t.Fatal("distribution should include routing subsection")
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
	end := strings.Index(out, "6. GOVERNANCE")
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
		Validators:     []model.Validator{{Moniker: "node1", Operator: "cosmosvaloper1abc"}},
		CommunityTax:   "2%",
		CommunityPool:  "0.5 PMT",
	}
	out := Build(d)
	for _, want := range []string{
		`class="dash-heading">1. STAKING</h2>`,
		`class="dash-heading">2. SLASHING</h2>`,
		`class="dash-heading">3. DISTRIBUTION</h2>`,
		`class="dash-heading">4. FEE MARKET</h2>`,
		`class="dash-heading">5. REWARDS</h2>`,
		`class="dash-heading">6. GOVERNANCE</h2>`,
		`class="dash-heading">1. INFRASTRUCTURE</h2>`,
		`class="dash-heading">2. VALIDATOR</h2>`,
		`class="dash-heading">3. EVM JSON-RPC</h2>`,
		`class="dash-subheading">For operators</h3>`,
		`class="dash-subheading">Probe health</h3>`,
		`class="kpi-grid"`,
		`class="kpi-tile"`,
		`class="data-table"`,
		`val-summary--p2p`,
		`class="dash-layer__title">Validator set</h3>`,
		"Block reward ledger",
		`class="dash-subheading">Routing</h3>`,
		"eco-domain--pmtrewards",
		`dash-section--rewards`,
		`dash-section--distribution`,
		`dash-section--feemarket`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("HTML inventory missing %q", want)
		}
	}
}
