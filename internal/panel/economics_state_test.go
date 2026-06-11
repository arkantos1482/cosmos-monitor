package panel

import (
	"strings"
	"testing"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func TestEconomicsInactivePMTDisabled(t *testing.T) {
	d := model.Report{
		PMTEnabled:      false,
		Inflation:       0,
		CommunityTax:    "2.00%",
		CommunityTaxPct: 2,
	}
	out := BuildView(ViewEconomics, d)
	if !strings.Contains(out, `eco-row--inactive`) {
		t.Fatal("expected inactive ledger rows when PMT disabled and inflation off")
	}
	if !strings.Contains(out, `eco-domain--pmtrewards`) {
		t.Fatal("expected PMT Rewards source card")
	}
	if !strings.Contains(out, `badge--bad">false`) {
		t.Fatal("expected false badge for disabled PMT")
	}
	if !strings.Contains(out, `eco-domain__row--inactive`) {
		t.Fatal("expected inactive inflation in domain card")
	}
	if strings.Contains(out, `id="eco-flags"`) {
		t.Fatal("flags panel should be removed")
	}
}

func TestEconomicsPMTPoolEmptyWarn(t *testing.T) {
	d := model.Report{
		PMTEnabled:        true,
		PMTPoolEmpty:      true,
		PMTRate:           "0.1 PMT/block",
		Inflation:         3.5,
		InflationPerBlock: "0.01 PMT/block",
		CommunityTax:      "2.00%",
		CommunityTaxPct:   2,
		BondedCount:       4,
		Validators:        []model.Validator{{CommissionFloat: 10}},
	}
	out := BuildView(ViewEconomics, d)
	if !strings.Contains(out, `eco-row--warn`) {
		t.Fatal("expected warn styling for empty PMT pool")
	}
	if !strings.Contains(out, `pool empty`) {
		t.Fatal("expected pool empty check in ledger")
	}
}

func TestEconomicsLedgerStep1NoPoolBalanceColumn(t *testing.T) {
	d := model.Report{
		PMTEnabled: true,
		PMTRate:    "0.1 PMT/block",
		PMTBalance: "1.00M PMT",
		BondedCount: 4,
		Validators:  []model.Validator{{CommissionFloat: 10}},
	}
	rows := economicsLedgerRows(d)
	if len(rows) == 0 || rows[0].Cells[1] != "x/pmtrewards → fee_collector" {
		t.Fatal("expected PMT ledger row first")
	}
	if rows[0].Cells[3] != "—" {
		t.Fatalf("step 1 balance now should be —, got %q", rows[0].Cells[3])
	}
}

func TestEconomicsCommunityTaxZeroInactive(t *testing.T) {
	d := model.Report{
		PMTEnabled:       true,
		PMTRate:          "0.1 PMT/block",
		PMTBalance:       "1M PMT",
		CommunityTax:     "0%",
		CommunityTaxZero: true,
		CommunityTaxPct:  0,
		BondedCount:      4,
		Validators:       []model.Validator{{CommissionFloat: 10}},
	}
	rows := economicsLedgerRows(d)
	var taxRow *EcoLedgerRow
	for i := range rows {
		if len(rows[i].Cells) > 1 && rows[i].Cells[1] == "community tax → pool" {
			taxRow = &rows[i]
			break
		}
	}
	if taxRow == nil {
		t.Fatal("missing community tax row")
	}
	if !taxRow.Inactive {
		t.Fatal("community tax row should be inactive when tax is 0%")
	}
}

func TestStakingCardModuleAccounts(t *testing.T) {
	bech := "cosmos1akkvh0ahmve830rj4mhkdnqs49kzw23c63nhdx"
	wantEVM := "0xEDACCBBFB7DB3278BC72AEEF66CC10A96C272A38"
	d := model.Report{
		BondedPct: 55.5,
		ModuleAccounts: []model.ModuleAccountRow{
			{Name: "bonded_tokens_pool", Address: bech, Balance: "10M PMT"},
		},
	}
	card := stakingCardHTML(d, false)
	if !strings.Contains(card, "bonded_tokens_pool") {
		t.Fatal("staking card should show bonded_tokens_pool")
	}
	if !strings.Contains(card, wantEVM) {
		t.Fatalf("staking card should show EVM address %q", wantEVM)
	}
}

func TestStakingCardNoGoalText(t *testing.T) {
	d := model.Report{
		BondedPct:  55.5,
		GoalBonded: 67,
		BondedAmt:  "10M PMT",
	}
	card := stakingCardHTML(d, false)
	if strings.Contains(strings.ToLower(card), "goal") {
		t.Fatal("staking card must not contain goal text")
	}
	if !strings.Contains(card, "55.50%") {
		t.Fatal("staking card should show bonded percentage")
	}
	if strings.Contains(card, "slash") {
		t.Fatal("staking card must not contain slashing params")
	}
}

func TestSlashingCardSeparate(t *testing.T) {
	d := model.Report{
		SlashWindow:   "10000",
		MinSigned:     50,
		SlashDowntime: "0.01%",
		SlashDS:       "5%",
	}
	card := slashingCardHTML(d, false)
	if !strings.Contains(card, `eco-domain--slashing`) {
		t.Fatal("expected slashing domain card")
	}
	if !strings.Contains(card, "signed blocks window") {
		t.Fatal("slashing card should show signed blocks window")
	}
	if !strings.Contains(card, "slash / downtime") {
		t.Fatal("slashing card should show downtime slash")
	}
}
