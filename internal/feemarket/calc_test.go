package feemarket

import (
	"testing"

	sdkmath "cosmossdk.io/math"
)

func TestCalcGasBaseFeeIncrease(t *testing.T) {
	parent := sdkmath.LegacyNewDec(1_000_000_000)
	got := CalcGasBaseFee(60_000_000, 50_000_000, 8, parent, MinUnitGas, sdkmath.LegacyZeroDec())
	want := sdkmath.LegacyNewDec(1_025_000_000)
	if !got.Equal(want) {
		t.Fatalf("increase: got %s want %s", got, want)
	}
}

func TestCalcGasBaseFeeDecrease(t *testing.T) {
	parent := sdkmath.LegacyNewDec(1_000_000_000)
	got := CalcGasBaseFee(40_000_000, 50_000_000, 8, parent, MinUnitGas, sdkmath.LegacyZeroDec())
	want := sdkmath.LegacyNewDec(975_000_000)
	if !got.Equal(want) {
		t.Fatalf("decrease: got %s want %s", got, want)
	}
}

func TestCalcGasBaseFeeStable(t *testing.T) {
	parent := sdkmath.LegacyNewDec(1_000_000_000)
	got := CalcGasBaseFee(50_000_000, 50_000_000, 8, parent, MinUnitGas, sdkmath.LegacyZeroDec())
	if !got.Equal(parent) {
		t.Fatalf("stable: got %s want %s", got, parent)
	}
}

func TestCalcGasBaseFeeNilMinGasPrice(t *testing.T) {
	parent := sdkmath.LegacyNewDec(1_000_000_000)
	got := CalcGasBaseFee(40_000_000, 50_000_000, 8, parent, MinUnitGas, sdkmath.LegacyDec{})
	if got.IsNil() {
		t.Fatal("expected valid decrease result")
	}
}
