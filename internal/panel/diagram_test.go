package panel

import (
	"strings"
	"testing"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func TestEconomicsOpDelPerBlockEdges(t *testing.T) {
	d := model.Report{
		PMTEnabled:  true,
		PMTRate:     "0.1 PMT/block",
		BondedCount: 4,
		Validators: []model.Validator{
			{CommissionFloat: 10}, {CommissionFloat: 10},
			{CommissionFloat: 10}, {CommissionFloat: 10},
		},
	}
	src := economicsOverviewMermaid(d)
	// 0.1/block, 4 equal vals → 0.025/val; 10% comm → ~0.0025 op
	if !strings.Contains(src, "~0.0025 PMT/blk") {
		t.Fatalf("expected per-block commission edge; got excerpt:\n%s", src)
	}
}

func TestEconomicsOverviewMermaidTopology(t *testing.T) {
	d := model.Report{
		Inflation:  3.5,
		PMTEnabled: true,
		PMTRate:    "0.1 PMT/block",
	}
	src := economicsOverviewMermaid(d)
	if !strings.Contains(src, "fee_collector") {
		t.Fatal("expected fee_collector node")
	}
	if !strings.Contains(src, "pmtPool -->") || !strings.Contains(src, "fc") {
		t.Fatal("expected PMT pool → fee_collector path")
	}
}

func TestFeemarketMechanicsNoDistribution(t *testing.T) {
	src := feemarketMechanicsMermaid(model.Report{
		BaseFee: "1000", Elasticity: 2, BlockGas: "21000",
		ParentBlockGasWanted: 21000,
	})
	for _, forbidden := range []string{"x/distribution", "Community pool", "fee_collector"} {
		if strings.Contains(src, forbidden) {
			t.Fatalf("feemarket diagram must not mention payout path: %q", forbidden)
		}
	}
	if !strings.HasPrefix(strings.TrimSpace(src), "graph LR") {
		t.Fatal("feemarket diagram should be LR")
	}
}
