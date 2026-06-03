package main

import (
	"strings"
	"testing"
)

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
	for _, src := range []string{
		economicsOverviewMermaid(d),
		feemarketMechanicsMermaid(d),
	} {
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

func TestEconomicsOverviewNoInflationNode(t *testing.T) {
	src := economicsOverviewMermaid(WebData{Inflation: 0, PMTEnabled: false})
	if strings.Contains(src, "infl[") {
		t.Fatal("inflation node should be omitted when rate is 0")
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
