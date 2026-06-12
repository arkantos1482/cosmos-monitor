package panel

import (
	"strings"
	"testing"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

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
		SlashWindow:    "10,000",
		SlashMaxMissed: 5000,
		MinSigned:      50,
		DowntimeJail:   "1m",
		SlashDowntime:  "0.01%",
		SlashDS:        "5%",
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
	if !strings.Contains(card, `data-table--penalties`) {
		t.Fatal("slashing card should show penalty matrix")
	}
	if !strings.Contains(card, "miss &gt; 5,000 / 10,000 window") {
		t.Fatal("downtime trigger should use live window thresholds")
	}
	if !strings.Contains(card, "permanent") {
		t.Fatal("double-sign row should show permanent jail")
	}
	if !strings.Contains(card, "penalty-tag--param") {
		t.Fatal("penalty cells should use neutral param styling")
	}
	if strings.Contains(card, "penalty-tag--slash") {
		t.Fatal("penalty cells should not use alarming slash styling")
	}
}
