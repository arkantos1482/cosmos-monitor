package panel

import (
	"strings"
	"testing"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func TestRewardsSectionConsolidatesChainAndLocal(t *testing.T) {
	d := model.Report{
		Inflation:           3.5,
		InflationPerBlock:   "0.01 PMT/block",
		PMTEnabled:          true,
		PMTRate:             "0.1000 PMT/block",
		PMTBalance:          "1.00M PMT",
		LastBlockFees:       "0.001 PMT  _(parent block gas × base fee)_",
		CommunityTaxPct:     2,
		BondedCount:         4,
		Validators:          []model.Validator{{CommissionFloat: 10}},
		Local: model.LocalValidator{
			IsValidator:      true,
			Moniker:          "node1",
			Commission:       10,
			VPPercent:        25,
			Outstanding:      "0.001 PMT",
			CommissionEarned: "0.0001 PMT",
		},
	}
	out := BuildView(ViewRewards, d)
	for _, want := range []string{
		`class="dash-heading">4. REWARDS</h2>`,
		`class="dash-subheading">This validator</h3>`,
		`class="eco-domain__title">PMT Rewards`,
		`class="eco-domain__title">Inflation`,
		"eco-domain--pmtrewards",
		"eco-domain--inflation",
		"per-block commission",
		"per-block delegators",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("rewards view missing %q", want)
		}
	}
	localIdx := strings.Index(out, `class="dash-subheading">This validator</h3>`)
	pmtIdx := strings.Index(out, `class="eco-domain__title">PMT Rewards`)
	inflIdx := strings.Index(out, `class="eco-domain__title">Inflation`)
	if localIdx < 0 || pmtIdx < 0 || inflIdx < 0 || localIdx > pmtIdx || pmtIdx > inflIdx {
		t.Fatal("rewards section must be ordered This validator → PMT Rewards → Inflation")
	}
	if strings.Count(out, `class="eco-domain__title">PMT Rewards`) != 1 {
		t.Fatal("rewards section must contain exactly one PMT Rewards block")
	}
	if strings.Count(out, `class="eco-domain__title">Inflation`) != 1 {
		t.Fatal("rewards section must contain exactly one Inflation block")
	}
	for _, gone := range []string{
		`class="dash-layer__title">Network-wide</h3>`,
		"eco-domain--blockrewards",
	} {
		if strings.Contains(out, gone) {
			t.Fatalf("rewards view should not contain %q", gone)
		}
	}
}

func TestLocalValidatorRewardsMovedFromNodeSection(t *testing.T) {
	d := model.Report{
		PMTEnabled:      true,
		PMTRate:         "0.1000 PMT/block",
		CommunityTaxPct: 2,
		BondedCount:     4,
		Validators:      []model.Validator{{CommissionFloat: 10}},
		Local: model.LocalValidator{
			IsValidator:      true,
			Moniker:          "node1",
			Commission:       10,
			VPPercent:        25,
			Outstanding:      "0.001 PMT",
			CommissionEarned: "0.0001 PMT",
		},
	}
	nodeOut := BuildView(ViewNode, d)
	for _, gone := range []string{
		`class="dash-subheading">Rewards</h3>`,
		"per-block commission",
		"outstanding rewards",
	} {
		if strings.Contains(nodeOut, gone) {
			t.Fatalf("validator section should not contain %q", gone)
		}
	}
}
