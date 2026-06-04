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
		BlockHeight: "100", BaseFee: "0.001 PMT", BaseFeeRaw: "1000000000000",
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
	if !strings.Contains(ex.Receipt, "Recorded demand") {
		t.Fatal("missing receipt demand line")
	}
	if len(ex.ParamRows) == 0 {
		t.Fatal("expected param rows")
	}
	if ex.StatBaseFee == "" {
		t.Fatal("expected stat base fee")
	}
}

func TestBuildFeemarketExplainNoBaseFee(t *testing.T) {
	ex := buildFeemarketExplain(model.Report{NoBaseFee: true, BaseFee: "1 PMT"})
	if !ex.NoBaseFee || ex.TrafficLabel != "FIXED PRICING" {
		t.Fatalf("no_base_fee mode: %+v", ex)
	}
}
