package main

import (
	"strings"
	"testing"

	"cosmossdk.io/math"
)

func TestCalcGasBaseFeeDecrease(t *testing.T) {
	parent := math.LegacyNewDec(1000)
	minUnit := math.LegacyOneDec()
	minGas := math.LegacyZeroDec()
	got := calcGasBaseFee(21000, 50_000_000, 8, parent, minUnit, minGas)
	if !got.LT(parent) {
		t.Fatalf("expected decrease, got %s", got)
	}
}

func TestInferParentBaseFeeRoundTrip(t *testing.T) {
	parent := math.LegacyNewDec(1_000_000_000_000)
	minUnit := math.LegacyOneDec()
	minGas := math.LegacyZeroDec()
	wanted, target := uint64(21_000), uint64(50_000_000)
	denom := uint64(8)
	current := calcGasBaseFee(wanted, target, denom, parent, minUnit, minGas)
	inf, ok := inferParentBaseFee(current, wanted, target, denom, minUnit, minGas)
	if !ok {
		t.Fatal("infer failed")
	}
	if !inf.Equal(parent) {
		t.Fatalf("parent %s inferred %s", parent, inf)
	}
}

func TestKatexTextLitScientificNotation(t *testing.T) {
	got := katexTextLit("7.00e-18 PMT")
	if !strings.Contains(got, `\text{7.00e-18 PMT}`) {
		t.Fatalf("expected text literal wrapper, got %q", got)
	}
}

func TestBuildFeemarketExplainKatexAndReceipt(t *testing.T) {
	d := WebData{
		BlockHeight:              "100",
		BaseFee:                  "0.001 PMT",
		BaseFeeRaw:               "1000000000000",
		BlockGas:                 "21000",
		BlockGasLimit:            100_000_000,
		Elasticity:               2,
		BaseFeeChangeDenominator: 8,
		MinGasMultiplier:         "0.5",
		ParentBlockGasUsed:       21000,
		ParentBlockGasWanted:     21000,
		ParentBlockResultsOK:     true,
		MinGasPriceRaw:           "0",
		EVMDenom:                 "apmt",
	}
	ex := buildFeemarketExplain(d)
	if !strings.Contains(ex.LatexGeneral, `W_{\text{stored}}`) {
		t.Fatal("missing general latex")
	}
	if !strings.Contains(ex.LatexSubstituted, `\[`) {
		t.Fatal("substituted latex must use display math delimiters")
	}
	if strings.Contains(ex.LatexSubstituted, `\begin{aligned}`) {
		t.Fatal("substituted latex must not use aligned (& breaks HTML)")
	}
	if !strings.Contains(ex.LatexSubstituted, "21{,}000") && !strings.Contains(ex.LatexSubstituted, "21000") {
		t.Fatal("missing substituted values")
	}
	if !strings.Contains(ex.LatexSubstituted, `\text{`) {
		t.Fatal("fee amounts must be in \\text{} for KaTeX")
	}
	if strings.Contains(ex.LatexSubstituted, "21,000") {
		t.Fatal("commas in math mode should use {,} separator")
	}
	if !strings.Contains(ex.TextReceipt, "CalcGasBaseFee") {
		t.Fatal("missing text receipt")
	}
}
