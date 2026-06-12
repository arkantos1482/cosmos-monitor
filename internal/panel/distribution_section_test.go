package panel

import (
	"strings"
	"testing"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func distributionChunk(t *testing.T, out string) string {
	t.Helper()
	idx := strings.Index(out, "5. DISTRIBUTION")
	end := strings.Index(out, `class="dash-heading">6. GOVERNANCE</h2>`)
	if idx < 0 || end < 0 {
		t.Fatal("expected distribution and governance sections")
	}
	return out[idx:end]
}

func TestWriteDistributionOverviewLedger(t *testing.T) {
	d := model.Report{
		Inflation:           0,
		InflationPerBlock:   "",
		PMTEnabled:          true,
		PMTRate:             "0.1000 PMT/block",
		PMTDailyEmit:        "~8640 PMT/day",
		PMTBalance:          "1.00M PMT",
		CommunityTax:        "2.00%",
		CommunityTaxPct:     2,
		BondedCount:         4,
		CommunityPool:       "0.50 PMT",
		UnclaimedDelegator:  "0.006169 PMT",
		UnclaimedCommission: "0.000685 PMT",
		TotalOutstanding:    "0.006854 PMT  across 4 validators",
		LastBlockFees:       "0.001 PMT  _(parent block gas × base fee)_",
		ModuleAccounts: []model.ModuleAccountRow{
			{Name: "fee_collector", Balance: "0 PMT", Role: "fees"},
			{Name: "distribution", Address: "cosmos1akkvh0ahmve830rj4mhkdnqs49kzw23c63nhdx", Balance: "0 PMT", Role: "escrow"},
			{Name: "bonded_tokens_pool", Balance: "12.5M PMT", Role: "staking"},
			{Name: "not_bonded_tokens_pool", Balance: "5.8M PMT", Role: "staking"},
		},
		Validators: []model.Validator{{CommissionFloat: 10}},
		Local: model.LocalValidator{
			IsValidator:      true,
			Moniker:          "node1",
			Commission:       10,
			VPPercent:        25,
			Outstanding:      "0.001 PMT",
			CommissionEarned: "0.0001 PMT",
		},
	}
	chunk := distributionChunk(t, Build(d))
	if strings.Contains(chunk, `class="diagram-panel mermaid"`) {
		t.Fatal("distribution section should not use mermaid")
	}

	for _, want := range []string{
		`class="dash-subheading">Routing</h3>`,
		"fee_collector",
		"distribution escrow",
		`class="eco-dist"`,
		`class="eco-acct__addr"`,
		"Unclaimed rewards",
		"delegator share",
		"validator commission",
		"community tax",
		"Block reward ledger",
		`class="dist-summary"`,
	} {
		if !strings.Contains(chunk, want) {
			t.Fatalf("distribution chunk missing %q", want)
		}
	}

	for _, gone := range []string{
		"eco-domain--pmtrewards",
		"eco-domain--inflation",
		"eco-domain--staking",
		"eco-domain--slashing",
		"bonded_tokens_pool",
		"not_bonded_tokens_pool",
		"Module accounts",
		"eco-module-accounts",
		"eco-domain--txfees",
		`class="dash-subheading">Chain parameters (reference)</h3>`,
		"eco-domain__hint",
		`class="dash-subheading">Advanced parameters (reward flow)</h3>`,
		`id="eco-flags"`,
		"eco-domain--distribution",
		"eco-domain--rewards",
		"At a glance",
		"Money flow (live balances)",
		"Distribution split",
		"Network total",
	} {
		if strings.Contains(chunk, gone) {
			t.Fatalf("distribution chunk should not contain %q", gone)
		}
	}
}

func TestDistributionSourcesProvenance(t *testing.T) {
	d := model.Report{
		PMTEnabled:          true,
		PMTRate:             "0.1 PMT/block",
		UnclaimedDelegator:  "0.006 PMT",
		UnclaimedCommission: "0.0006 PMT",
		ModuleAccounts: []model.ModuleAccountRow{
			{Name: "fee_collector", Balance: "0 PMT"},
		},
	}
	out := BuildWithOptions(d, Options{ShowSources: true})
	chunk := distributionChunk(t, out)
	for _, want := range []string{
		`class="dash-sources"`,
		`>Data sources</summary>`,
		`class="hint-provenance"`,
		"distribution/v1beta1/validators",
		"outstanding_rewards",
		"distribution/v1beta1/community_pool",
	} {
		if !strings.Contains(chunk, want) {
			t.Fatalf("distribution data sources missing %q", want)
		}
	}
	for _, gone := range []string{
		"eco-domain__hint",
		"pmtrewards/v1/params",
		"Tx fees:",
	} {
		if strings.Contains(chunk, gone) {
			t.Fatalf("distribution should not contain %q", gone)
		}
	}
}

func TestModuleAccountDisplayAddressInDistribution(t *testing.T) {
	bech := "cosmos1akkvh0ahmve830rj4mhkdnqs49kzw23c63nhdx"
	wantEVM := "0xEDACCBBFB7DB3278BC72AEEF66CC10A96C272A38"
	d := model.Report{
		ModuleAccounts: []model.ModuleAccountRow{
			{Name: "distribution", Address: bech, Balance: "0 PMT"},
		},
		CommunityTax: "2%",
	}
	out := BuildView(ViewDistribution, d)
	if strings.Contains(out, bech) {
		t.Fatal("distribution should prefer EVM hex over bech32 for module addresses")
	}
	if !strings.Contains(out, wantEVM) {
		t.Fatalf("distribution should show EVM address %q", wantEVM)
	}
}
