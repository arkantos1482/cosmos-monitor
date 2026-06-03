package main

import (
	"strings"
	"testing"
)

func TestMermaidDottedEdgeRenders(t *testing.T) {
	src := "graph TD\n  stake[x/staking]\n  dist[x/distribution]\n  stake -.->|voting power| dist\n"
	out, err := renderMermaid(src)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "staking") || !strings.Contains(out, "distribution") {
		t.Fatalf("expected both nodes in render:\n%s", out)
	}
}

func TestRenderMermaidEconomicsOverview(t *testing.T) {
	d := WebData{
		Inflation:        3.5,
		BondedPct:        72.3,
		GoalBonded:       67,
		CommunityTax:     "2.00%",
		CommunityTaxPct:  2,
		PMTEnabled:       true,
		PMTRate:          "0.1000 PMT/block",
		BaseFee:          "0.000000000000000000",
		GasPrice:         "0",
		BlockGas:         "21000",
		Elasticity:       2,
		BaseFeeChangeDenominator: 8,
	}
	out, err := renderMermaid(economicsOverviewMermaidASCII(d))
	if err != nil {
		t.Fatalf("render economics: %v", err)
	}
	if !strings.Contains(out, "┌") && !strings.Contains(out, "+") {
		t.Fatalf("expected box drawing, got:\n%s", out)
	}
	for _, src := range []string{feemarketMechanicsMermaid(d)} {
		out, err := renderMermaid(src)
		if err != nil {
			t.Fatalf("render: %v\nsource:\n%s", err, src)
		}
		if !strings.Contains(out, "┌") && !strings.Contains(out, "+") {
			t.Fatalf("expected box drawing, got:\n%s", out)
		}
	}
}

func TestEconomicsOverviewMermaidTopology(t *testing.T) {
	d := WebData{
		Inflation:  3.5,
		PMTEnabled: true,
		PMTRate:    "0.1 PMT/block",
	}
	src := economicsOverviewMermaid(d)
	if !strings.Contains(src, "fee_collector") {
		t.Fatal("expected fee_collector node")
	}
	if strings.Contains(src, "dist --> pmt") || strings.Contains(src, "dist-->pmt") {
		t.Fatal("must not route distribution to PMT pool")
	}
	if !strings.Contains(src, "pmtPool -->") || !strings.Contains(src, "fc") {
		t.Fatal("expected PMT pool → fee_collector path")
	}
	if !strings.Contains(src, "fees --> fc") {
		t.Fatal("expected fees → fee_collector path")
	}
}

func TestEconomicsOverviewASCIINoSubgraph(t *testing.T) {
	src := economicsOverviewMermaidASCII(WebData{Inflation: 0, GoalBonded: 67, PMTEnabled: true, PMTRate: "0.1 PMT/block"})
	if strings.Contains(src, "subgraph") {
		t.Fatal("ASCII economics diagram must not use subgraphs")
	}
	if !strings.HasPrefix(strings.TrimSpace(src), "graph TD") {
		t.Fatal("ASCII economics diagram should be top-down")
	}
}

func TestEconomicsOverviewZeroInflationShowsMint(t *testing.T) {
	src := economicsOverviewMermaid(WebData{Inflation: 0, GoalBonded: 67, PMTEnabled: false})
	if !strings.Contains(src, "infl[") {
		t.Fatal("x/mint node should always be present")
	}
	if !strings.Contains(src, "0% inflation") {
		t.Fatal("expected inactive inflation label")
	}
	if !strings.Contains(src, "goal bonded 67%") {
		t.Fatal("goal bonded belongs on x/mint, not distribution")
	}
	for _, line := range strings.Split(src, "\n") {
		if strings.Contains(line, "dist[") && strings.Contains(line, "goal bonded") {
			t.Fatal("goal bonded must not appear on distribution node")
		}
	}
	if !strings.Contains(src, "infl -->") {
		t.Fatal("expected mint → fee_collector edge")
	}
	if !strings.Contains(src, "subgraph sources") {
		t.Fatal("expected inflows subgraph for layout")
	}
	if strings.Contains(src, "subgraph modules") {
		t.Fatal("modules subgraph breaks mermaid-ascii layout")
	}
	if strings.Contains(src, "stake -.->") || strings.Contains(src, "stake[") {
		t.Fatal("use staking node with solid voting-power edge")
	}
	if strings.Contains(src, "op --> del") {
		t.Fatal("operator and delegators are parallel splits from validators, not sequential")
	}
	if !strings.Contains(src, "val -->|commission") || !strings.Contains(src, "val -->|remainder") {
		t.Fatal("expected parallel commission/remainder edges from validators")
	}
}

func TestFeemarketMechanicsNoDistribution(t *testing.T) {
	src := feemarketMechanicsMermaid(WebData{BaseFee: "1000", Elasticity: 2})
	for _, forbidden := range []string{"x/distribution", "Community pool", "Validators + delegators", "pmtPool", "fee_collector"} {
		if strings.Contains(src, forbidden) {
			t.Fatalf("feemarket diagram must not mention payout path: %q", forbidden)
		}
	}
	if !strings.Contains(src, "CalculateBaseFee") {
		t.Fatal("expected CalculateBaseFee in feemarket diagram")
	}
}
