package panel

import (
	"strings"
	"testing"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func TestEconomicsDistributionEscrowMerged(t *testing.T) {
	wantEVM := "0xEDACCBBFB7DB3278BC72AEEF66CC10A96C272A38"
	d := model.Report{
		ModuleAccounts: []model.ModuleAccountRow{
			{Name: "distribution", Address: "cosmos1akkvh0ahmve830rj4mhkdnqs49kzw23c63nhdx", Balance: "0 PMT"},
		},
	}
	html := economicsDistItemsHTML([]economicsDistItem{{
		param:   "distribution escrow",
		balance: "0 PMT",
		addr:    economicsDistributionModuleAddr(d),
		effect:  "x/distribution module escrow",
	}})
	for _, want := range []string{
		`class="eco-dist"`,
		`class="eco-acct"`,
		`class="eco-acct__balance"`,
		`class="eco-acct__addr"`,
		wantEVM,
		"0 PMT",
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("distribution escrow HTML missing %q:\n%s", want, html)
		}
	}
}

func TestEconomicsUnclaimedRewardsWithAddress(t *testing.T) {
	wantEVM := "0xEDACCBBFB7DB3278BC72AEEF66CC10A96C272A38"
	d := model.Report{
		UnclaimedDelegator:  "0.006169 PMT",
		UnclaimedCommission: "0.000685 PMT",
		TotalOutstanding:    "0.006854 PMT  across 4 validators",
		ModuleAccounts: []model.ModuleAccountRow{
			{Name: "distribution", Address: "cosmos1akkvh0ahmve830rj4mhkdnqs49kzw23c63nhdx", Balance: "0 PMT"},
		},
	}
	out := BuildView(ViewEconomics, d)
	idx := strings.Index(out, "Unclaimed rewards")
	if idx < 0 {
		t.Fatal("missing Unclaimed rewards subsection")
	}
	chunk := out[idx:]
	end := strings.Index(chunk, `class="dash-subheading">`)
	if end > 0 {
		chunk = chunk[:end]
	}
	for _, want := range []string{
		"delegator share",
		"validator commission",
		"total outstanding",
		`class="eco-acct__addr"`,
		wantEVM,
	} {
		if !strings.Contains(chunk, want) {
			t.Fatalf("unclaimed rewards chunk missing %q", want)
		}
	}
}
