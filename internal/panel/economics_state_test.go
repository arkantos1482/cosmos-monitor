package panel

import (
	"strings"
	"testing"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func TestEconomicsInactivePMTDisabled(t *testing.T) {
	d := model.Report{
		PMTEnabled:    false,
		Inflation:     0,
		CommunityTax:  "2.00%",
		CommunityTaxPct: 2,
	}
	out := BuildView(ViewEconomics, d)
	if !strings.Contains(out, `eco-row--inactive`) {
		t.Fatal("expected inactive ledger rows when PMT disabled and inflation off")
	}
	if !strings.Contains(out, `pmtrewards.enabled`) {
		t.Fatal("expected flags panel with pmtrewards.enabled")
	}
	if !strings.Contains(out, `badge--bad">false`) {
		t.Fatal("expected false badge for disabled PMT")
	}
	if !strings.Contains(out, `badge--bad">inflation off`) {
		t.Fatal("expected inflation off badge in summary")
	}
}

func TestEconomicsPMTPoolEmptyWarn(t *testing.T) {
	d := model.Report{
		PMTEnabled:   true,
		PMTPoolEmpty: true,
		PMTRate:      "0.1 PMT/block",
		Inflation:    3.5,
		InflationPerBlock: "0.01 PMT/block",
		CommunityTax: "2.00%",
		CommunityTaxPct: 2,
		BondedCount:  4,
		Validators:   []model.Validator{{CommissionFloat: 10}},
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
