package main

import (
	"strings"
	"testing"
)

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
	if !strings.Contains(src, `infl -->|"mint off (0%)"|`) && !strings.Contains(src, "infl -->|mint off (0%)|") {
		t.Fatal("expected dynamic mint edge label")
	}
	if !strings.Contains(src, "infl -->") {
		t.Fatal("expected mint → fee_collector edge")
	}
	if !strings.Contains(src, "subgraph sources") {
		t.Fatal("expected inflows subgraph for layout")
	}
	if strings.Contains(src, "subgraph modules") {
		t.Fatal("unexpected modules subgraph")
	}
	if strings.Contains(src, "stake -.->") || strings.Contains(src, "stake[") {
		t.Fatal("use staking node with solid voting-power edge")
	}
	if strings.Contains(src, "op --> del") {
		t.Fatal("operator and delegators are parallel splits from validators, not sequential")
	}
	if !strings.Contains(src, "val -->|") || !strings.Contains(src, "commission") || !strings.Contains(src, "remainder") {
		t.Fatal("expected parallel commission/remainder edges from validators")
	}
}

func TestFeemarketMechanicsNoDistribution(t *testing.T) {
	src := feemarketMechanicsMermaid(WebData{
		BaseFee: "1000", Elasticity: 2, BlockGas: "21000",
		ParentBlockGasWanted: 21000,
	})
	for _, forbidden := range []string{"x/distribution", "Community pool", "Validators + delegators", "pmtPool", "fee_collector"} {
		if strings.Contains(src, forbidden) {
			t.Fatalf("feemarket diagram must not mention payout path: %q", forbidden)
		}
	}
	for _, want := range []string{"CalculateBaseFee", "stored W", "EndBlock N-1"} {
		if !strings.Contains(src, want) {
			t.Fatalf("expected %q in feemarket diagram", want)
		}
	}
	if !strings.HasPrefix(strings.TrimSpace(src), "graph LR") {
		t.Fatal("feemarket diagram should be LR")
	}
}
