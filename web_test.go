package main

import (
	"strings"
	"testing"
)

func TestBuildMarkdownMermaidFence(t *testing.T) {
	d := WebData{Inflation: 3.5, PMTEnabled: true, PMTRate: "0.1 PMT/block"}
	md := buildMarkdown(d)
	if !strings.Contains(md, "```mermaid") {
		t.Fatal("markdown should use mermaid fenced blocks")
	}
	if strings.Contains(md, `<div class="mermaid">`) {
		t.Fatal("markdown should not embed raw mermaid HTML")
	}
	if !strings.Contains(md, "graph LR") {
		t.Fatal("economics diagram should use LR layout")
	}
}

func TestBuildMarkdownFeeMath(t *testing.T) {
	d := WebData{
		BlockHeight: "100", BaseFee: "1", BaseFeeRaw: "1000",
		BlockGas: "21000", ParentBlockGasWanted: 21000,
		BlockGasLimit: 100_000_000, Elasticity: 2,
		BaseFeeChangeDenominator: 8, MinGasMultiplier: "0.5",
		ParentBlockResultsOK: true,
	}
	md := buildMarkdown(d)
	if !strings.Contains(md, "\n$$\n") {
		t.Fatal("fee math should use $$ display blocks")
	}
	if strings.Contains(md, `class="fee-math-tex"`) {
		t.Fatal("markdown should not use HTML katex hooks")
	}
}

func TestRenderFragmentMermaid(t *testing.T) {
	d := WebData{Inflation: 3.5, PMTEnabled: true, PMTRate: "0.1 PMT/block", GoalBonded: 67}
	out := renderFragment(d)
	if !strings.Contains(out, "language-mermaid") && !strings.Contains(out, "mermaid") {
		t.Fatal("rendered fragment should include mermaid source")
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
		t.Fatal("node labels must use <br/> not literal newlines")
	}
	if !strings.Contains(src, "<br/>") {
		t.Fatal("node labels should use <br/> line breaks")
	}
	if !strings.Contains(src, `|"block 482,160"|`) {
		t.Fatal("edge labels with commas must be quoted for mermaid.js")
	}
	if strings.Contains(src, `-->|block 482,160|`) {
		t.Fatal("unquoted comma edge label breaks mermaid.js")
	}
}
