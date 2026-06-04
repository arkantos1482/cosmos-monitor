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

func TestLatexTextLitScientificNotation(t *testing.T) {
	got := latexTextLit("7.00e-18 PMT")
	if !strings.Contains(got, `\text{7.00e-18 PMT}`) {
		t.Fatalf("expected text literal wrapper, got %q", got)
	}
}

func TestBuildFeemarketExplain(t *testing.T) {
	d := model.Report{
		BlockHeight: "100", BaseFee: "0.001 PMT", BaseFeeRaw: "1000000000000",
		BlockGas: "21000", BlockGasLimit: 100_000_000, Elasticity: 2,
		BaseFeeChangeDenominator: 8, MinGasMultiplier: "0.5",
		ParentBlockGasUsed: 21000, ParentBlockGasWanted: 21000,
		ParentBlockResultsOK: true, MinGasPriceRaw: "0", EVMDenom: "apmt",
	}
	ex := buildFeemarketExplain(d)
	if !strings.Contains(ex.LatexGeneral, `W_{\text{stored}}`) {
		t.Fatal("missing general latex")
	}
	if !strings.Contains(ex.LatexSubstituted, `\[`) {
		t.Fatal("substituted latex must use display math delimiters")
	}
	if !strings.Contains(ex.TextReceipt, "CalcGasBaseFee") {
		t.Fatal("missing text receipt")
	}
}
