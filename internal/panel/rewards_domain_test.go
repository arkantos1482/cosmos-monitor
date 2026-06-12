package panel

import (
	"strings"
	"testing"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func TestRewardsSectionConsolidatesChainAndLocal(t *testing.T) {
	d := model.Report{
		Inflation:         3.5,
		InflationPerBlock: "0.01 PMT/block",
		PMTEnabled:        true,
		PMTRate:           "0.1000 PMT/block",
		PMTBalance:        "1.00M PMT",
		CommunityTaxPct:   2,
		BondedCount:       4,
		Validators:        []model.Validator{{CommissionFloat: 10}},
		Local: model.LocalValidator{
			IsValidator: true,
			Moniker:     "node1",
			Commission:  10,
			VPPercent:   25,
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
		"rewards-summary",
		"per-block commission",
		"per-block delegators",
		`class="dash-subheading">Emission sources</h3>`,
		`data-table--emission`,
		"reward_per_block",
		"annual_provisions",
		"x/mint MintFn",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("rewards view missing %q", want)
		}
	}
	localIdx := strings.Index(out, `class="dash-subheading">This validator</h3>`)
	pmtIdx := strings.Index(out, `class="eco-domain__title">PMT Rewards`)
	emIdx := strings.Index(out, `class="dash-subheading">Emission sources</h3>`)
	if localIdx < 0 || pmtIdx < 0 || emIdx < 0 || localIdx > pmtIdx || pmtIdx > emIdx {
		t.Fatal("rewards section must be ordered This validator → domain cards → Emission sources")
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
			IsValidator: true,
			Moniker:     "node1",
			Commission:  10,
			VPPercent:   25,
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

func TestRewardsInactivePMTDisabled(t *testing.T) {
	d := model.Report{PMTEnabled: false, Inflation: 0}
	out := BuildView(ViewRewards, d)
	for _, want := range []string{
		`eco-domain--pmtrewards eco-domain--inactive`,
		`eco-domain__status badge badge--bad">disabled`,
		`badge--bad">false`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("rewards view missing %q", want)
		}
	}
}

func TestRewardsPMTPoolEmptyWarn(t *testing.T) {
	d := model.Report{
		PMTEnabled:        true,
		PMTPoolEmpty:      true,
		PMTRate:           "0.1 PMT/block",
		Inflation:         3.5,
		InflationPerBlock: "0.01 PMT/block",
	}
	out := BuildView(ViewRewards, d)
	if !strings.Contains(out, `eco-domain--pmtrewards eco-domain--ineffective`) {
		t.Fatal("expected ineffective PMT card when pool empty")
	}
	if !strings.Contains(out, `eco-domain__status badge badge--warn">pool empty`) {
		t.Fatal("expected pool empty status badge")
	}
}

func TestInflationCardInactiveWhenZero(t *testing.T) {
	d := model.Report{Inflation: 0}
	card := mintInflationDomainCard(d)
	for _, want := range []string{
		`eco-domain--inflation eco-domain--inactive`,
		`eco-domain__status badge badge--bad">off`,
	} {
		if !strings.Contains(card, want) {
			t.Fatalf("inflation card missing %q:\n%s", want, card)
		}
	}
}

func TestInflationCardActiveWhenMinting(t *testing.T) {
	d := model.Report{Inflation: 5, InflationPerBlock: "0.01 PMT/block", AnnualProvisions: "1M PMT"}
	card := mintInflationDomainCard(d)
	if !strings.Contains(card, `eco-domain__status badge badge--ok">minting`) {
		t.Fatalf("expected minting badge:\n%s", card)
	}
}

func TestPMTRewardsPoolMerged(t *testing.T) {
	wantEVM := "0xEDACCBBFB7DB3278BC72AEEF66CC10A96C272A38"
	d := model.Report{
		PMTEnabled:     true,
		PMTRate:        "0.1 PMT/block",
		PMTBalance:     "1.00M PMT",
		PMTPoolAddress: "cosmos1akkvh0ahmve830rj4mhkdnqs49kzw23c63nhdx",
	}
	card := pmtRewardsDomainCard(d)
	for _, want := range []string{
		"pool_address",
		`class="eco-acct"`,
		wantEVM,
		"1.00M PMT",
		"NewPMTRewardMintFn",
	} {
		if !strings.Contains(card, want) {
			t.Fatalf("PMT rewards card missing %q:\n%s", want, card)
		}
	}
}

func TestRewardsEmissionPerBlockCombined(t *testing.T) {
	d := model.Report{
		PMTEnabled:        true,
		PMTRate:           "0.1 PMT/block",
		InflationPerBlock: "0.01 PMT/block",
	}
	got := rewardsEmissionPerBlock(d)
	if got == "—" || !strings.Contains(got, "/block") {
		t.Fatalf("rewardsEmissionPerBlock = %q", got)
	}
}
