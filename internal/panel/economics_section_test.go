package panel

import (
	"strings"
	"testing"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func TestWriteEconomicsOverviewLedger(t *testing.T) {
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
	out := Build(d)
	idx := strings.Index(out, "3. ECONOMICS")
	end := strings.Index(out, `class="dash-heading">4. FEE MARKET</h2>`)
	if idx < 0 || end < 0 {
		t.Fatal("expected economics and fee market sections")
	}
	chunk := out[idx:end]
	if strings.Contains(chunk, `class="diagram-panel mermaid"`) {
		t.Fatal("economics section should not use mermaid")
	}

	for _, want := range []string{
		`class="dash-subheading">Distribution</h3>`,
		"fee_collector",
		"distribution escrow",
		`class="eco-dist"`,
		`class="eco-acct__addr"`,
		"Unclaimed rewards",
		"delegator share",
		"validator commission",
		"community tax",
	} {
		if !strings.Contains(chunk, want) {
			t.Fatalf("economics chunk missing %q", want)
		}
	}

	for _, gone := range []string{
		"Block reward ledger",
		"eco-domains",
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
		"Total/block",
		"per-block commission",
	} {
		if strings.Contains(chunk, gone) {
			t.Fatalf("economics chunk should not contain %q", gone)
		}
	}
	
	// Check that old KPI + badges layout is gone
	if strings.Contains(chunk, "At a glance") {
		t.Fatal("economics should not duplicate at-a-glance subsection")
	}
	for _, gone := range []string{
		"Money flow (live balances)",
		"This validator", 
		"this validator → commission",
		"your commission",
		"Distribution split",
		"Network total",
		"per-block commission",
	} {
		if strings.Contains(chunk, gone) {
			t.Fatalf("economics chunk should not contain old table %q", gone)
		}
	}
	
}

func TestEconomicsPerBlockSplit(t *testing.T) {
	d := model.Report{
		PMTEnabled:      true,
		PMTRate:         "0.1 PMT/block",
		CommunityTaxPct: 2,
		BondedCount:     4,
		Validators:      []model.Validator{{CommissionFloat: 10}},
	}
	tax, valPool, op, del, unit, ok := economicsPerBlockSplit(d)
	if !ok || unit != "PMT" {
		t.Fatalf("split not ok: ok=%v unit=%q", ok, unit)
	}
	if tax < 0.00199 || tax > 0.00201 {
		t.Fatalf("tax want ~0.002 got %v", tax)
	}
	if valPool < 0.0979 || valPool > 0.0981 {
		t.Fatalf("valPool want ~0.098 got %v", valPool)
	}
	if op < 0.0024 || op > 0.0026 {
		t.Fatalf("op want ~0.0025 got %v", op)
	}
	if del < 0.0220 || del > 0.0230 {
		t.Fatalf("del want ~0.022 got %v", del)
	}
}

func TestFeeCollectorBalanceAndChecks(t *testing.T) {
	d := model.Report{
		ModuleAccounts: []model.ModuleAccountRow{
			{Name: "fee_collector", Balance: "0 PMT"},
		},
		UnclaimedDelegator:  "0.006 PMT",
		UnclaimedCommission: "0.0006 PMT",
		TotalOutstanding:    "0.0066 PMT  across 4 validators",
	}
	if got := FeeCollectorBalance(d); got != "0 PMT" {
		t.Fatalf("FeeCollectorBalance = %q", got)
	}
	if economicsFeeCollectorCheck(d) != "cleared" {
		t.Fatalf("expected cleared, got %q", economicsFeeCollectorCheck(d))
	}
	if economicsUnclaimedCheck(d) != "sums match" {
		t.Fatalf("expected sums match, got %q", economicsUnclaimedCheck(d))
	}
}

func TestEconomicsSourcesProvenance(t *testing.T) {
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
	idx := strings.Index(out, "3. ECONOMICS")
	end := strings.Index(out, `class="dash-heading">4. FEE MARKET</h2>`)
	if idx < 0 || end < 0 {
		t.Fatal("expected economics section")
	}
	chunk := out[idx:end]
	for _, want := range []string{
		`class="dash-sources"`,
		`>Data sources</summary>`,
		`class="hint-provenance"`,
		"distribution/v1beta1/validators",
		"outstanding_rewards",
		"distribution/v1beta1/community_pool",
	} {
		if !strings.Contains(chunk, want) {
			t.Fatalf("economics data sources missing %q", want)
		}
	}
	for _, gone := range []string{
		"pmtrewards/v1/params",
	} {
		if strings.Contains(chunk, gone) {
			t.Fatalf("economics should not contain rewards source %q", gone)
		}
	}
	for _, gone := range []string{
		"eco-domain__hint",
		"Tx fees:",
	} {
		if strings.Contains(chunk, gone) {
			t.Fatalf("economics should not contain %q", gone)
		}
	}
}

func TestModuleAccountDisplayAddress(t *testing.T) {
	bech := "cosmos1akkvh0ahmve830rj4mhkdnqs49kzw23c63nhdx"
	wantEVM := "0xEDACCBBFB7DB3278BC72AEEF66CC10A96C272A38"
	d := model.Report{
		ModuleAccounts: []model.ModuleAccountRow{
			{Name: "fee_collector", Address: bech, Balance: "0 PMT"},
		},
	}
	if got := moduleAccountDisplayAddress(d, "fee_collector"); got != wantEVM {
		t.Fatalf("moduleAccountDisplayAddress = %q want %q", got, wantEVM)
	}
	out := Build(d)
	if strings.Contains(out, bech) {
		t.Fatal("economics should prefer EVM hex over bech32 for module addresses")
	}
	if !strings.Contains(out, wantEVM) {
		t.Fatalf("economics should show EVM address %q", wantEVM)
	}
}

func TestEconomicsPerBlockSplitInflationOnly(t *testing.T) {
	d := model.Report{
		Inflation:         5,
		InflationPerBlock: "0.01 PMT/block",
		CommunityTaxPct:   2,
		BondedCount:       4,
		Validators:        []model.Validator{{CommissionFloat: 10}},
	}
	tax, valPool, _, _, unit, ok := economicsPerBlockSplit(d)
	if !ok || unit != "apmt" {
		t.Fatalf("split not ok: ok=%v unit=%q", ok, unit)
	}
	if tax < 0.00019 || tax > 0.00021 {
		t.Fatalf("tax want ~0.0002 got %v", tax)
	}
	if valPool < 0.0097 || valPool > 0.0099 {
		t.Fatalf("valPool want ~0.0098 got %v", valPool)
	}
}

func TestRewardInPerBlockTotal(t *testing.T) {
	d := model.Report{
		PMTEnabled:        true,
		PMTRate:           "0.1 PMT/block",
		InflationPerBlock: "0.01 PMT/block",
		LastBlockFees:     "0.001 PMT  _(parent block gas × base fee)_",
	}
	got := RewardInPerBlockTotal(d)
	if !strings.Contains(got, "/block") {
		t.Fatalf("RewardInPerBlockTotal = %q", got)
	}
}
