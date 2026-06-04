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
	if !strings.Contains(ex.HeroLine, "485,592") || !strings.Contains(ex.HeroLine, "50,000,000") {
		t.Fatalf("hero line missing block/target: %q", ex.HeroLine)
	}
	if len(ex.LastBlockRows) < 4 {
		t.Fatalf("expected last-block rows, got %d", len(ex.LastBlockRows))
	}
	if !strings.Contains(ex.LastBlockRows[1][0], "gas_wanted") {
		t.Fatalf("missing gas_wanted row: %v", ex.LastBlockRows[1])
	}
	if !strings.Contains(ex.FormulaLine, "50,000,000") || strings.Contains(ex.FormulaLine, "MaxUint64") {
		t.Fatalf("finite chain formula should use real target, not sentinel: %q", ex.FormulaLine)
	}
	if len(ex.ThisBlockRows) == 0 || ex.ThisBlockRows[0][0] != "base_fee" {
		t.Fatalf("this-block rows: %v", ex.ThisBlockRows)
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
	if !strings.Contains(ex.WalletLine, "0.001 PMT") {
		t.Fatalf("wallet line should include base fee: %q", ex.WalletLine)
	}
	if !strings.Contains(ex.ChainLine, "50,000,000") {
		t.Fatalf("chain line should include target: %q", ex.ChainLine)
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
	combined := ex.HeroLine + ex.FormulaLine + ex.LastBlockRows[2][1]
	for _, s := range []string{"MaxUint64", "unlimited", "sentinel"} {
		if !strings.Contains(combined, s) {
			t.Fatalf("expected %q in explain output", s)
		}
	}
	for _, row := range ex.LastBlockRows {
		for _, cell := range row {
			if strings.Contains(cell, mystery) {
				t.Fatalf("last block row should not show mystery number: %v", row)
			}
		}
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
	if ex.FormulaLine != "" {
		t.Fatalf("no formula in no_base_fee mode: %q", ex.FormulaLine)
	}
}
