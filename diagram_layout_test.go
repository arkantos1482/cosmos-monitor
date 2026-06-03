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
		TotalOutstanding: "1.2 PMT",
		PMTEnabled:       true,
		PMTRate:          "0.1 PMT/block",
		PMTBalance:       "500K PMT",
		Validators: []WebValidator{
			{CommissionFloat: 10},
			{CommissionFloat: 10},
		},
	}
	src := economicsOverviewMermaid(d)
	for _, want := range []string{"mempool 2", "400M PMT", "4 validators", "0 PMT", "500K PMT", "0.1 PMT/block"} {
		if !strings.Contains(src, want) {
			t.Fatalf("expected live label fragment %q in:\n%s", want, src)
		}
	}
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
