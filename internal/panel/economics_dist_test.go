package panel

import (
	"strings"
	"testing"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func TestDistributionCardCommunityTax(t *testing.T) {
	d := model.Report{
		CommunityTax:        "2.00%",
		CommunityPool:       "0.50 PMT",
		CommunityTaxPct:     2,
		WithdrawAddrEnabled: true,
		UnclaimedDelegator:  "0.006169 PMT",
		UnclaimedCommission: "0.000685 PMT",
		ModuleAccounts: []model.ModuleAccountRow{
			{Name: "distribution", Address: "cosmos1akkvh0ahmve830rj4mhkdnqs49kzw23c63nhdx", Balance: "0 PMT"},
		},
	}
	html := distributionCardHTML(d)
	for _, want := range []string{
		`eco-domain--distribution`,
		"community_tax",
		"2.00%",
		"withdraw_addr_enabled",
		"community pool",
		"0.50 PMT",
		"Community treasury",
		"Distribution escrow",
		"for delegators",
		"for operators",
		"total unclaimed",
		"escrow check",
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("distribution card missing %q:\n%s", want, html)
		}
	}
}

func TestDistributionUnclaimedTotal(t *testing.T) {
	d := model.Report{
		UnclaimedDelegator:  "0.006169 PMT",
		UnclaimedCommission: "0.000685 PMT",
	}
	total := distributionUnclaimedTotal(d)
	if total == "" {
		t.Fatal("expected unclaimed total")
	}
}
