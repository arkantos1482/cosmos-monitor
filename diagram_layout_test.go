package main

import (
	"strings"
	"testing"
)

func TestEconomicsOverviewLiveLabels(t *testing.T) {
	d := WebData{
		MempoolTxs:       2,
		BondedCount:      4,
		BondedAmt:        "400M PMT",
		CommunityPool:    "0 PMT",
		CommunityTaxZero: true,
		TotalOutstanding: "1.2 PMT  across 4 validators",
		PMTEnabled:       true,
		PMTRate:          "0.1 PMT/block",
		PMTBalance:       "500K PMT",
		Validators: []WebValidator{
			{CommissionFloat: 10},
			{CommissionFloat: 10},
		},
	}
	src := economicsOverviewMermaid(d)
	for _, want := range []string{"mempool 2", "400M PMT", "4 validators", "0 PMT", "500K PMT", "0.1 PMT/block", "across 4 validators", "x/staking", "voting power"} {
		if !strings.Contains(src, want) {
			t.Fatalf("expected live label fragment %q in:\n%s", want, src)
		}
	}
	if !strings.Contains(src, "fees --> fc") || strings.Contains(src, "dist --> pmt") {
		t.Fatalf("unexpected topology in:\n%s", src)
	}
}

func TestEconomicsOverviewStackedLabelsRender(t *testing.T) {
	d := WebData{
		BondedCount:      4,
		BondedAmt:        "400.00M PMT",
		TotalOutstanding: "0.006854 PMT  across 4 validators",
		PMTEnabled:       true,
		PMTRate:          "0.1 PMT/block",
		PMTPoolEmpty:     true,
	}
	out, err := renderMermaid(economicsOverviewMermaidASCII(d))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"fee_collector", "unclaimed 0.006854 PMT", "across 4 validators", "400.00M PMT bonded", "0% inflation", "x/mint BeginBlock", "x/staking"} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected %q in render output", want)
		}
	}
	h := strings.Count(out, "\n") + 1
	w := diagramMaxWidth(out)
	t.Logf("economics diagram: %dx%d", w, h)
	if w > 410 {
		t.Fatalf("diagram too wide for one page: %d cols", w)
	}
	if h > 45 {
		t.Fatalf("diagram too tall for one page: %d lines", h)
	}
	for _, bad := range []string{"validator┴", "st┴king", "dist┴"} {
		if strings.Contains(out, bad) {
			t.Fatalf("ASCII label/box collision %q in:\n%s", bad, out)
		}
	}
}

func TestMultilineLabelNarrowsBox(t *testing.T) {
	flat := mermaidLabel("fee_collector · outstanding 0.006854 PMT  across 4 validators")
	stacked := stackMermaidQuoted(stackLabelText(
		"fee_collector",
		"outstanding 0.006854 PMT",
		"across 4 validators",
	))
	srcFlat := "graph TD\n  fc[" + flat + "]\n"
	srcStack := "graph TD\n  fc[" + stacked + "]\n"
	outFlat, err := renderMermaid(srcFlat)
	if err != nil {
		t.Fatal(err)
	}
	outStack, err := renderMermaid(srcStack)
	if err != nil {
		t.Fatal(err)
	}
	maxFlat, maxStack := diagramMaxWidth(outFlat), diagramMaxWidth(outStack)
	if maxStack >= maxFlat {
		t.Fatalf("expected stacked label to narrow box: flat=%d stack=%d\nflat:\n%s\nstack:\n%s",
			maxFlat, maxStack, outFlat, outStack)
	}
}

func diagramMaxWidth(s string) int {
	m := 0
	for _, l := range strings.Split(s, "\n") {
		if len(l) > m {
			m = len(l)
		}
	}
	return m
}

func TestFeemarketMechanicsVerticalSpine(t *testing.T) {
	src := feemarketMechanicsMermaid(WebData{
		BlockGas:                 "21000",
		Elasticity:               2,
		BaseFee:                  "1000",
		BaseFeeChangeDenominator: 8,
		MinGasPrice:              "0.001 PMT",
		AdjCap:                   "±0.01/block",
	})
	// No fan-in: calc should have a single predecessor in the chain.
	if strings.Contains(src, "gasUsed --> calc\n") {
		t.Fatal("expected vertical chain, not gasUsed --> calc fan-in")
	}
	if !strings.Contains(src, "gasUsed --> gasTarget\n") {
		t.Fatal("expected vertical spine starting gasUsed --> gasTarget")
	}
	for _, want := range []string{"gas used: 21000", "elasticity 2", "change denom 8", "min_gas 0.001 PMT", "max Δ"} {
		if !strings.Contains(src, want) {
			t.Fatalf("expected value %q preserved", want)
		}
	}
}
