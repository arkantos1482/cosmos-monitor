package feemarket

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func TestCalcGasBaseFeeDecrease(t *testing.T) {
	parent := math.LegacyNewDec(1000)
	got := CalcGasBaseFee(21000, 50_000_000, 8, parent, MinUnitGas, math.LegacyZeroDec())
	if !got.LT(parent) {
		t.Fatalf("expected decrease, got %s", got)
	}
}

func TestCalcGasBaseFeeAtFloor(t *testing.T) {
	parent := math.LegacyNewDec(7)
	maxStep := parent.QuoInt(math.NewIntFromUint64(8)).TruncateDec()
	if !maxStep.IsZero() {
		t.Fatalf("expected zero max decrease step at base 7, got %s", maxStep)
	}
	c := LoadContext(model.Report{
		BlockHeight: "100", BaseFeeRaw: "7",
		BlockGasLimit: 100_000_000, Elasticity: 2,
		BaseFeeChangeDenominator: 8,
		ParentBlockGasWanted: 21_000,
		MinGasPriceRaw: "0",
	})
	if c.Badge.Label != "AT FLOOR" {
		t.Fatalf("badge: got %q want AT FLOOR", c.Badge.Label)
	}
}

func TestLoadContextBadgeAtFloor(t *testing.T) {
	c := LoadContext(model.Report{
		BlockHeight: "100", BaseFee: "7 apmt", BaseFeeRaw: "7",
		BlockGasLimit: ^uint64(0), Elasticity: 2,
		BaseFeeChangeDenominator: 8,
		ParentBlockGasWanted: 2_000_000, ParentBlockGasUsed: 2_000_000,
		MinGasPriceRaw: "0",
	})
	if c.Badge.Label != "AT FLOOR" {
		t.Fatalf("badge: got %q want AT FLOOR", c.Badge.Label)
	}
	if c.Badge.Label == "FALLING" {
		t.Fatal("AT FLOOR and FALLING must be mutually exclusive")
	}
}

func TestLoadContextBadgeFalling(t *testing.T) {
	c := LoadContext(model.Report{
		BlockHeight: "100", BaseFeeRaw: "1200000000",
		BlockGasLimit: 30_000_000, Elasticity: 2,
		BaseFeeChangeDenominator: 8,
		ParentBlockGasWanted: 12_400_000, ParentBlockGasUsed: 12_400_000,
		MinGasPriceRaw: "0",
	})
	if c.Badge.Label != "FALLING" {
		t.Fatalf("badge: got %q want FALLING", c.Badge.Label)
	}
	if c.DecreaseStep.IsNil() || !c.DecreaseStep.IsPositive() {
		t.Fatalf("expected positive decrease step, got %v", c.DecreaseStep)
	}
}

func TestLoadContextFeesDisabled(t *testing.T) {
	c := LoadContext(model.Report{
		BlockHeight: "50", EnableHeight: 100,
		BlockGasLimit: 30_000_000, Elasticity: 2,
	})
	if !c.FeesDisabled || c.Badge.Label != "FEES DISABLED" {
		t.Fatalf("expected FEES DISABLED, got %+v", c.Badge)
	}
}

func TestLoadContextUnlimitedMaxGas(t *testing.T) {
	c := LoadContext(model.Report{
		BlockHeight: "200", BaseFeeRaw: "1000",
		BlockGasLimit: ^uint64(0), Elasticity: 2,
		BaseFeeChangeDenominator: 8,
		ParentBlockGasWanted: 21000,
	})
	if !c.UnlimitedBlockGas {
		t.Fatal("expected unlimited max_gas")
	}
	if c.UtilPct != "" {
		t.Fatal("sentinel target should not produce utilization pct")
	}
}

func TestLoadContextFiniteMaxGas(t *testing.T) {
	c := LoadContext(model.Report{
		BlockHeight: "100", BaseFeeRaw: "1000",
		BlockGasLimit: 100_000_000, Elasticity: 2,
		BaseFeeChangeDenominator: 8,
		ParentBlockGasWanted: 21000,
	})
	if c.Target != 50_000_000 {
		t.Fatalf("target: got %d", c.Target)
	}
	if c.UtilPct == "" {
		t.Fatal("finite max_gas should produce utilization")
	}
}
