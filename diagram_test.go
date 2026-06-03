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
		PMTRate:          "1000apmt",
		BaseFee:          "1000000000",
		GasPrice:         "20000000000",
	}
	for _, src := range []string{
		economicsOverviewMermaid(d),
		feeFlowMermaid(d),
		pmtRewardsMermaid(d),
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

func TestPMTRewardsMermaidDisabled(t *testing.T) {
	if pmtRewardsMermaid(WebData{PMTEnabled: false}) != "" {
		t.Fatal("expected empty when PMT disabled")
	}
}
