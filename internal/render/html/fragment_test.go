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
	if !strings.Contains(out, `class="fee-traffic"`) {
		t.Fatal("fragment should include fee-market traffic panel")
	}
}

func TestRenderFragmentEconomicsTables(t *testing.T) {
	d := model.Report{
		Inflation: 3.5, PMTEnabled: true, PMTRate: "0.1 PMT/block", GoalBonded: 67,
		ModuleAccounts: []model.ModuleAccountRow{{Name: "fee_collector", Balance: "1 PMT"}},
	}
	out := RenderFragment(d)
	if !strings.Contains(out, "Block reward ledger") {
		t.Fatal("rendered fragment should include economics ledger")
	}
	if !strings.Contains(out, `class="dash-subheading">Chain parameters (reference)</h3>`) {
		t.Fatal("rendered fragment should include economics reference subsection")
	}
	if strings.Contains(out, `data-details-key="economics-chain-params"`) {
		t.Fatal("economics reference should not use collapsible details")
	}
}
