package panel

import (
	"strings"
	"testing"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func rewardsChunk(t *testing.T, out string) string {
	t.Helper()
	idx := strings.Index(out, "5. REWARDS")
	end := strings.Index(out, `class="dash-heading">6. GOVERNANCE</h2>`)
	if idx < 0 || end < 0 {
		t.Fatal("expected rewards and governance sections")
	}
	return out[idx:end]
}

func TestWriteRewardsSectionSourcesOnly(t *testing.T) {
	d := model.Report{
		PMTEnabled:        true,
		PMTRate:           "0.1000 PMT/block",
		Inflation:         3.5,
		InflationPerBlock: "0.01 PMT/block",
		Local: model.LocalValidator{
			IsValidator: true,
			Commission:  10,
			VPPercent:   25,
		},
		Validators:      []model.Validator{{CommissionFloat: 10}},
		CommunityTaxPct: 2,
		BondedCount:     4,
	}
	chunk := rewardsChunk(t, Build(d))
	for _, want := range []string{
		"eco-domain--pmtrewards",
		"eco-domain--inflation",
		`class="eco-summary"`,
		"per-block commission",
		"per-block delegators",
	} {
		if !strings.Contains(chunk, want) {
			t.Fatalf("rewards chunk missing %q", want)
		}
	}
	for _, gone := range []string{
		"Block reward ledger",
		`class="dash-subheading">Routing</h3>`,
		"Unclaimed rewards",
		"community tax",
		`class="eco-dist"`,
	} {
		if strings.Contains(chunk, gone) {
			t.Fatalf("rewards chunk should not contain %q", gone)
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
