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
		`class="dash-heading">2. REWARDS</h2>`,
		`class="dash-layer__title">Chain</h3>`,
		`class="dash-layer__title">This validator</h3>`,
		"eco-domain--pmtrewards",
		"eco-domain--inflation",
		"Block reward ledger",
		`class="dash-subheading">Unclaimed</h3>`,
		"outstanding rewards",
		"commission earned",
		"per-block commission",
		"per-block delegators",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("rewards view missing %q", want)
		}
	}
	for _, gone := range []string{
		`class="dash-subheading">Distribution</h3>`,
		"Unclaimed rewards",
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

func TestRewardsSourcesProvenance(t *testing.T) {
	d := model.Report{
		PMTEnabled: true,
		PMTRate:    "0.1 PMT/block",
	}
	out := BuildViewWithOptions(ViewRewards, d, Options{ShowSources: true})
	for _, want := range []string{
		`class="dash-sources"`,
		"pmtrewards/v1/params",
		"mint/v1beta1/inflation",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("rewards data sources missing %q", want)
		}
	}
	for _, gone := range []string{
		"distribution/v1beta1/community_pool",
	} {
		if strings.Contains(out, gone) {
			t.Fatalf("rewards should not contain economics source %q", gone)
		}
	}
}
