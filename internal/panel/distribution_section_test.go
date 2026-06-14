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

func TestWriteDistributionSection(t *testing.T) {
	d := model.Report{
		CommunityTax:        "2.00%",
		CommunityTaxPct:     2,
		WithdrawAddrEnabled: true,
		BondedCount:         4,
		CommunityPool:       "0.50 PMT",
		UnclaimedDelegator:  "0.006169 PMT",
		UnclaimedCommission: "0.000685 PMT",
		ModuleAccounts: []model.ModuleAccountRow{
			{Name: "distribution", Address: "cosmos1akkvh0ahmve830rj4mhkdnqs49kzw23c63nhdx", Balance: "0 PMT", Role: "escrow"},
		},
		Validators: []model.Validator{{
			Moniker: "node1", Operator: "cosmosvaloper1abc",
			Outstanding: "0.001 PMT", CommissionEarned: "0.0001 PMT", CommissionFloat: 10,
		}},
		Local: model.LocalValidator{
			IsValidator:      true,
			Moniker:          "node1",
			Commission:       10,
			Outstanding:      "0.001 PMT",
			CommissionEarned: "0.0001 PMT",
		},
	}
	chunk := distributionChunk(t, Build(d))

	for _, want := range []string{
		`class="dist-summary"`,
		`dist-summary__scope-label">This validator`,
		`dist-summary__scope-label">Network`,
		`unclaimed-stack__head-label">unclaimed total`,
		`class="unclaimed-stack"`,
		`unclaimed-stack__op`,
		`eco-domain--distribution`,
		`class="eco-domain__title">Distribution`,
		"community_tax",
		"withdraw_addr_enabled",
		"community pool",
		"Community treasury",
		"Withdraw policy",
		`class="dist-escrow"`,
		"distribution escrow",
		`class="dash-subheading">This validator</h3>`,
		`class="dash-subheading">Unclaimed rewards</h3>`,
		`class="dash-subheading">Treasury &amp; params</h3>`,
		"delegator share",
		`data-table--staking-set`,
		`table-scroll--fit`,
		">total<",
		">commission<",
		">outstanding share<",
	} {
		if !strings.Contains(chunk, want) {
			t.Fatalf("distribution chunk missing %q", want)
		}
	}

	for _, gone := range []string{
		"Block reward ledger",
		`class="dash-subheading">Routing</h3>`,
		`class="dash-subheading">Network-wide</h3>`,
		`data-table--ledger`,
		"eco-domain--pmtrewards",
		"eco-domain--inflation",
		"eco-domain--staking",
		"bonded_tokens_pool",
		`class="eco-dist"`,
		"At a glance",
		"Money flow",
	} {
		if strings.Contains(chunk, gone) {
			t.Fatalf("distribution chunk should not contain %q", gone)
		}
	}
}

func TestDistributionSourcesProvenance(t *testing.T) {
	d := model.Report{
		CommunityTax:        "2%",
		WithdrawAddrEnabled: true,
		UnclaimedDelegator:  "0.006 PMT",
		UnclaimedCommission: "0.0006 PMT",
		ModuleAccounts: []model.ModuleAccountRow{
			{Name: "fee_collector", Balance: "0 PMT"},
		},
		Exchanges: append(sampleExchanges(), []model.SourceExchange{
			{
				Kind: "http", Method: "GET",
				URL:      "http://localhost:1317/cosmos/staking/v1beta1/validators?status=BOND_STATUS_BONDED",
				Request:  "(none)",
				Response: `{"validators":[]}`,
				OK:       true, Latency: "4ms",
			},
			{
				Kind: "http", Method: "GET",
				URL:      "http://localhost:1317/cosmos/distribution/v1beta1/community_pool",
				Request:  "(none)",
				Response: `{"pool":[]}`,
				OK:       true, Latency: "5ms",
			},
		}...),
	}
	out := BuildWithOptions(d, Options{ShowSources: true})
	chunk := distributionChunk(t, out)
	for _, want := range []string{
		`class="dash-sources"`,
		`>Data sources</summary>`,
		`dash-sources__exchange`,
		`dash-sources__tag">req`,
		`dash-sources__tag">res`,
		`class="hint-provenance"`,
		"outstanding_rewards",
		"distribution/v1beta1/params",
		"distribution/v1beta1/community_pool",
		"staking/v1beta1/validators",
		"/status",
	} {
		if !strings.Contains(chunk, want) {
			t.Fatalf("distribution data sources missing %q", want)
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
		CommunityPool: "0.5 PMT",
	}
	out := BuildView(ViewDistribution, d)
	if strings.Contains(out, bech) {
		t.Fatal("distribution should prefer EVM hex over bech32 for module addresses")
	}
	if !strings.Contains(out, wantEVM) {
		t.Fatalf("distribution should show EVM address %q", wantEVM)
	}
}
