package html

import (
	"strings"
	"testing"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func TestStripDisplayMathForGoldmark(t *testing.T) {
	md := "intro\n\n$$\na+b\n$$\n\nafter"
	stripped, blocks := stripDisplayMathForGoldmark(md)
	if len(blocks) != 1 || blocks[0] != "a+b" {
		t.Fatalf("blocks=%#v", blocks)
	}
	if !strings.Contains(stripped, "PMTOP_MATH_BLOCK_0") {
		t.Fatalf("stripped=%q", stripped)
	}
}

func TestRenderFragmentMathDisplay(t *testing.T) {
	d := model.Report{
		BlockHeight: "100", BaseFee: "1", BaseFeeRaw: "1000",
		BlockGas: "21000", ParentBlockGasWanted: 21000,
		BlockGasLimit: 100_000_000, Elasticity: 2,
		BaseFeeChangeDenominator: 8, MinGasMultiplier: "0.5",
		ParentBlockResultsOK: true,
	}
	out := RenderFragment(d)
	if !strings.Contains(out, `class="math-display"`) {
		t.Fatal("fragment should inject math-display nodes for KaTeX")
	}
}

func TestRenderFragmentMermaid(t *testing.T) {
	d := model.Report{Inflation: 3.5, PMTEnabled: true, PMTRate: "0.1 PMT/block", GoalBonded: 67}
	out := RenderFragment(d)
	if !strings.Contains(out, "mermaid") {
		t.Fatal("rendered fragment should include mermaid source")
	}
}
