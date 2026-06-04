package main

import (
	"strings"
	"testing"
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

func TestInjectDisplayMathHTML(t *testing.T) {
	html := "<p>PMTOP_MATH_BLOCK_0</p>"
	out := injectDisplayMathHTML(html, []string{`W_{\text{stored}} = 1`})
	if !strings.Contains(out, `class="math-display"`) || !strings.Contains(out, `data-tex-b64=`) {
		t.Fatalf("got %q", out)
	}
	if strings.Contains(out, "PMTOP_MATH_BLOCK") {
		t.Fatal("placeholder should be replaced")
	}
}

func TestRenderFragmentMathDisplay(t *testing.T) {
	d := WebData{
		BlockHeight: "100", BaseFee: "1", BaseFeeRaw: "1000",
		BlockGas: "21000", ParentBlockGasWanted: 21000,
		BlockGasLimit: 100_000_000, Elasticity: 2,
		BaseFeeChangeDenominator: 8, MinGasMultiplier: "0.5",
		ParentBlockResultsOK: true,
	}
	out := renderFragment(d)
	if !strings.Contains(out, `class="math-display"`) {
		t.Fatal("fragment should inject math-display nodes for KaTeX")
	}
	if strings.Contains(out, "<p>$$\n") {
		t.Fatal("raw broken $$ paragraphs should not appear in HTML")
	}
}
