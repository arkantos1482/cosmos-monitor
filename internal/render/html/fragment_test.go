package html

import (
	"strings"
	"testing"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func TestRenderFragmentFeeMarket(t *testing.T) {
	d := model.Report{
		BlockHeight: "100", BaseFee: "1", BaseFeeRaw: "1000",
		BlockGas: "21000", ParentBlockGasWanted: 21000,
		BlockGasLimit: 100_000_000, Elasticity: 2,
		BaseFeeChangeDenominator: 8, MinGasMultiplier: "0.5",
		ParentBlockResultsOK: true,
	}
	out := RenderFragment(d)
	if !strings.Contains(out, `id="fee-L1"`) {
		t.Fatal("fragment should include fee market L1 panel")
	}
}

func TestRenderFragmentEconomicsTables(t *testing.T) {
	d := model.Report{
		Inflation: 3.5, PMTEnabled: true, PMTRate: "0.1 PMT/block", GoalBonded: 67,
		ModuleAccounts: []model.ModuleAccountRow{{Name: "fee_collector", Balance: "1 PMT"}},
	}
	out := RenderFragment(d)
	for _, want := range []string{
		`class="dash-heading">3. FEE MARKET</h2>`,
		`class="dash-heading">4. REWARDS</h2>`,
		`class="dash-heading">5. DISTRIBUTION</h2>`,
		"eco-domain--distribution",
		"eco-domain--pmtrewards",
		"eco-domain--inflation",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("rendered fragment missing %q", want)
		}
	}
	if strings.Contains(out, `class="dash-subheading">Chain parameters (reference)</h3>`) {
		t.Fatal("economics reference subsection should be removed")
	}
}
