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
			{Name: "distribution", Balance: "0 PMT", Role: "escrow"},
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
	idx := strings.Index(out, "5. ECONOMICS")
	end := strings.Index(out, "6. GOVERNANCE")
	if idx < 0 || end < 0 {
		t.Fatal("expected economics and governance sections")
	}
	chunk := out[idx:end]
	if strings.Contains(chunk, `class="diagram-panel mermaid"`) {
		t.Fatal("economics section should not use mermaid")
	}
	for _, want := range []string{
		"At a glance",
		"Block reward ledger",
		"fee_collector",
		"this validator → commission",
		`<details class="dash-details">`,
		"Chain parameters (reference)",
	} {
		if !strings.Contains(chunk, want) {
			t.Fatalf("economics chunk missing %q", want)
		}
	}
	for _, gone := range []string{
		"Money flow (live balances)",
		"This validator",
		"Module account",
		"Distribution split",
		"Network total",
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
