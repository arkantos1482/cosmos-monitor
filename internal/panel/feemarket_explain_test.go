package panel

import (
	"strings"
	"testing"

	"cosmossdk.io/math"
	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func TestCalcGasBaseFeeDecrease(t *testing.T) {
	parent := math.LegacyNewDec(1000)
	got := calcGasBaseFee(21000, 50_000_000, 8, parent, math.LegacyOneDec(), math.LegacyZeroDec())
	if !got.LT(parent) {
		t.Fatalf("expected decrease, got %s", got)
	}
}

func TestBuildFeemarketExplain(t *testing.T) {
	d := model.Report{
		BlockHeight: "485,592", BaseFee: "0.001 PMT", BaseFeeRaw: "1000000000000",
		BlockGas: "21000", BlockGasLimit: 100_000_000, Elasticity: 2,
		BaseFeeChangeDenominator: 8, MinGasMultiplier: "0.5",
		ParentBlockGasUsed: 21000, ParentBlockGasWanted: 21000,
		ParentBlockResultsOK: true, MinGasPriceRaw: "0", EVMDenom: "apmt",
		GasPrice: "1200000000000",
	}
	ex := buildFeemarketExplain(d)
	if ex.TrafficLabel != "FEE FALLING" {
		t.Fatalf("traffic: got %q", ex.TrafficLabel)
	}
	if ex.NextAdj != "↓" {
		t.Fatalf("next adj: got %q", ex.NextAdj)
	}
	if ex.HeroLine != "Block 485,592" {
		t.Fatalf("hero should be block height only: %q", ex.HeroLine)
	}
	if len(ex.VariableRows) < 4 {
		t.Fatalf("expected variable rows, got %d", len(ex.VariableRows))
	}
	if ex.VariableRows[0][0] != "W" || !strings.Contains(ex.VariableRows[0][1], "Stored gas wanted") {
		t.Fatalf("W row should define stored gas wanted: %v", ex.VariableRows[0])
	}
	if ex.VariableRows[0][2] != "21,000" {
		t.Fatalf("W live value: got %q", ex.VariableRows[0][2])
	}
	formula := strings.Join(ex.FormulaBlocks, "\n")
	if !strings.Contains(formula, "50,000,000") || strings.Contains(formula, "MaxUint64") {
		t.Fatalf("finite chain formula should use real target, not sentinel: %q", formula)
	}
	if !strings.Contains(formula, "→ ↓") {
		t.Fatalf("formula should show fee direction arrow: %q", formula)
	}
	if len(ex.ParamRows) == 0 {
		t.Fatal("expected param rows")
	}
	foundElasticity := false
	for _, row := range ex.ParamRows {
		if len(row) >= 1 && row[0] == "elasticity_multiplier" {
			foundElasticity = true
			if strings.Contains(row[2], "last block") || strings.Contains(row[2], "%") {
				t.Fatalf("elasticity meaning should be static: %q", row[2])
			}
		}
	}
	if !foundElasticity {
		t.Fatal("missing elasticity row")
	}
}

func TestBuildFeemarketExplainUnlimitedMaxGas(t *testing.T) {
	d := model.Report{
		BlockHeight: "100", BaseFee: "0.001 PMT", BaseFeeRaw: "1000000000000",
		BlockGas: "21000", BlockGasLimit: ^uint64(0), Elasticity: 2,
		BaseFeeChangeDenominator: 8,
		ParentBlockGasUsed: 21000, ParentBlockGasWanted: 21000,
		ParentBlockResultsOK: true, EVMDenom: "apmt",
	}
	ex := buildFeemarketExplain(d)
	if !ex.UnlimitedBlockGas || !ex.HideLoadMeter {
		t.Fatalf("expected unlimited max_gas mode: Unlimited=%v HideMeter=%v", ex.UnlimitedBlockGas, ex.HideLoadMeter)
	}
	mystery := "9,223,372,036,854,775,807"
	if strings.Contains(ex.HeroLine, mystery) {
		t.Fatalf("hero should not show raw sentinel decimal: %q", ex.HeroLine)
	}
	combined := ex.HeroLine + strings.Join(ex.FormulaBlocks, "")
	for _, row := range ex.VariableRows {
		for _, cell := range row {
			combined += cell
		}
	}
	for _, s := range []string{"MaxUint64", "sentinel"} {
		if !strings.Contains(combined, s) {
			t.Fatalf("expected %q in explain output", s)
		}
	}
	if strings.Contains(combined, mystery) {
		t.Fatalf("output should not show mystery decimal: %q", combined)
	}
}

func TestBuildFeemarketExplainNoBaseFee(t *testing.T) {
	ex := buildFeemarketExplain(model.Report{NoBaseFee: true, BaseFee: "1 PMT", BlockHeight: "1"})
	if !ex.NoBaseFee || ex.TrafficLabel != "FIXED PRICING" {
		t.Fatalf("no_base_fee mode: %+v", ex)
	}
	if !strings.Contains(ex.HeroLine, "no_base_fee") {
		t.Fatalf("hero should mention no_base_fee: %q", ex.HeroLine)
	}
	if len(ex.FormulaBlocks) != 0 {
		t.Fatalf("no formulas in no_base_fee mode: %v", ex.FormulaBlocks)
	}
}
