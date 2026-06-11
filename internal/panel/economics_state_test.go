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
}
