package panel

import (
	"strings"
	"testing"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func TestWriteEconomicsOverviewTables(t *testing.T) {
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
			{Name: "fee_collector", Balance: "0.10 PMT", Role: "fees"},
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
	if strings.Contains(out, `class="diagram-panel mermaid"`) {
		idx := strings.Index(out, "5. ECONOMICS")
		end := strings.Index(out, "6. GOVERNANCE")
		if end < 0 {
			end = len(out)
		}
		chunk := out[idx:end]
		if strings.Contains(chunk, `class="diagram-panel mermaid"`) {
			t.Fatal("economics section should not use mermaid")
		}
	}
	if !strings.Contains(out, "Money flow (live balances)") {
		t.Fatal("expected economics overview subsection")
	}
	if !strings.Contains(out, "fee_collector") {
		t.Fatal("expected module account table row")
	}
	if !strings.Contains(out, "unclaimed delegator rewards") {
		t.Fatal("expected network totals table")
	}
	if !strings.Contains(out, "This validator") {
		t.Fatal("expected local validator table")
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
