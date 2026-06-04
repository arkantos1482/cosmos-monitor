package main

import (
	"strings"
	"testing"
)

func TestBuildMarkdownWebMermaid(t *testing.T) {
	d := WebData{Inflation: 3.5, PMTEnabled: true, PMTRate: "0.1 PMT/block"}
	md := buildMarkdown(d, true)
	if !strings.Contains(md, `<div class="mermaid">`) {
		t.Fatal("web markdown should embed raw mermaid HTML")
	}
	if !strings.Contains(md, "graph LR") {
		t.Fatal("web economics diagram should use LR layout")
	}
	if strings.Contains(md, "```text") {
		t.Fatal("web markdown should not include ASCII diagram blocks")
	}
}

func TestBuildMarkdownWebFeeMath(t *testing.T) {
	d := WebData{
		BlockHeight: "100", BaseFee: "1", BaseFeeRaw: "1000",
		BlockGas: "21000", ParentBlockGasWanted: 21000,
		BlockGasLimit: 100_000_000, Elasticity: 2,
		BaseFeeChangeDenominator: 8, MinGasMultiplier: "0.5",
		ParentBlockResultsOK: true,
	}
	md := buildMarkdown(d, true)
	if !strings.Contains(md, `class="fee-math"`) {
		t.Fatal("web markdown should include fee-math div")
	}
	if !strings.Contains(md, `class="fee-math-tex"`) {
		t.Fatal("web markdown should include katex render nodes")
	}
}

func TestRenderFragmentFeeMathHTMLSafe(t *testing.T) {
	d := WebData{
		BlockHeight: "482,764", BaseFee: "7.00e-18 PMT", BaseFeeRaw: "7000000000000",
		BlockGas: "0", ParentBlockGasWanted: 0, ParentBlockGasUsed: 0,
		BlockGasLimit: ^uint64(0), Elasticity: 2,
		BaseFeeChangeDenominator: 8, MinGasMultiplier: "0.5",
		ParentBlockResultsOK: true,
	}
	out := renderFragment(d)
	if !strings.Contains(out, `class="fee-math-tex"`) || !strings.Contains(out, `data-tex-b64=`) {
		t.Fatal("fee math must use base64 katex nodes, not raw latex in HTML")
	}
	if strings.Contains(out, `[ \\textbf`) || strings.Contains(out, "[ \\textbf") {
		t.Fatal("raw display-math delimiters must not appear in HTML")
	}
}

func TestBuildMarkdownTerminalASCII(t *testing.T) {
	d := WebData{Inflation: 3.5, PMTEnabled: true, PMTRate: "0.1 PMT/block"}
	md := buildMarkdown(d, false)
	if strings.Contains(md, `<div class="mermaid">`) {
		t.Fatal("terminal markdown should not embed mermaid HTML")
	}
	if !strings.Contains(md, "```text") {
		t.Fatal("terminal markdown should include ASCII diagram blocks")
	}
}

func TestRenderFragmentMermaidHTML(t *testing.T) {
	d := WebData{Inflation: 3.5, PMTEnabled: true, PMTRate: "0.1 PMT/block", GoalBonded: 67}
	out := renderFragment(d)
	if !strings.Contains(out, `class="mermaid"`) {
		t.Fatal("rendered fragment should preserve mermaid div")
	}
	if !strings.Contains(out, "graph LR") {
		t.Fatal("expected LR mermaid source in fragment")
	}
}

func TestEconomicsOverviewWebMermaidSyntax(t *testing.T) {
	d := WebData{
		Inflation:        3.5,
		BondedPct:        72.3,
		BlockHeight:      "482,160",
		CommunityTax:     "2.00%",
		PMTEnabled:       true,
		PMTRate:          "0.1000 PMT/block",
		TotalOutstanding: "0.006854 PMT  across 4 validators",
		Validators:       []WebValidator{{CommissionFloat: 10}},
	}
	src := economicsOverviewMermaid(d)
	if strings.Contains(src, "\nheight ") || strings.Contains(src, "\nmempool ") {
		t.Fatal("web node labels must use <br/> not literal newlines")
	}
	if !strings.Contains(src, "<br/>") {
		t.Fatal("web node labels should use <br/> line breaks")
	}
	if !strings.Contains(src, `|"block 482,160"|`) {
		t.Fatal("edge labels with commas must be quoted for mermaid.js")
	}
	if strings.Contains(src, `-->|block 482,160|`) {
		t.Fatal("unquoted comma edge label breaks mermaid.js")
	}
}
