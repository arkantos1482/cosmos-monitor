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
	rewardsIdx := strings.Index(out, "4. REWARDS")
	end := strings.Index(out, "5. DISTRIBUTION")
	if rewardsIdx < 0 || end < 0 {
		t.Fatal("expected rewards and distribution sections")
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
		t.Fatal("rewards should use domain cards")
	}
	if !strings.Contains(rewards, `rewards-summary`) {
		t.Fatal("rewards should include embedded summary KPIs")
	}

	for _, want := range []string{
		`class="eco-domain__title">PMT Rewards`,
		`class="eco-domain__title">Inflation`,
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

	distIdx := strings.Index(out, "5. DISTRIBUTION")
	govIdx := strings.Index(out, `class="dash-heading">6. GOVERNANCE</h2>`)
	if distIdx < 0 || govIdx < 0 {
		t.Fatal("expected distribution and governance sections")
	}
	dist := out[distIdx:govIdx]
	if !strings.Contains(dist, `eco-domain--distribution`) {
		t.Fatal("distribution should include x/distribution domain card")
	}
	if strings.Contains(dist, `data-table--ledger`) || strings.Contains(dist, "Block reward ledger") {
		t.Fatal("distribution should not include block reward ledger")
	}
}

func TestBuildFeeMarketPanel(t *testing.T) {
	d := model.Report{
		BlockHeight: "100", BaseFee: "1", BaseFeeRaw: "1000",
		BlockGas: "21000", ParentBlockGasWanted: 21000,
		BlockGasLimit: 100_000_000, Elasticity: 2,
		BaseFeeChangeDenominator: 8, MinGasMultiplier: "0.5",
		ParentBlockResultsOK: true, EVMDenom: "apmt",
	}
	out := Build(d)
	if !strings.Contains(out, `class="fm-summary"`) {
		t.Fatal("fee market section should render summary panel")
	}
	idx := strings.Index(out, `class="dash-heading">3. FEE MARKET</h2>`)
	end := strings.Index(out, "4. REWARDS")
	if idx < 0 || end < 0 {
		t.Fatal("expected fee market and rewards sections")
	}
	fee := out[idx:end]
	if strings.Contains(fee, `id="fee-L1"`) {
		t.Fatal("fee market section should not use legacy L1 ladder")
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
		`class="dash-heading">3. FEE MARKET</h2>`,
		`class="dash-heading">4. REWARDS</h2>`,
		`class="dash-heading">5. DISTRIBUTION</h2>`,
		`class="dash-heading">6. GOVERNANCE</h2>`,
		`class="dash-heading">1. INFRASTRUCTURE</h2>`,
		`class="dash-heading">2. VALIDATOR</h2>`,
		`class="dash-heading">3. EVM JSON-RPC</h2>`,
		`class="dash-subheading">Method probes</h3>`,
		`eco-domain__divider">MetaMask custom network`,
		`class="kpi-grid"`,
		`class="kpi-tile"`,
		`class="data-table"`,
		`val-summary--p2p`,
		`class="dash-layer__title">Validator set</h3>`,
		`eco-domain--distribution`,
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
