package panel

import (
	"strings"
	"testing"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func TestLocalUnclaimedBreakdownHTML(t *testing.T) {
	html := localUnclaimedBreakdownHTML(model.LocalValidator{
		Outstanding:      "0.001 PMT",
		CommissionEarned: "0.0001 PMT",
	})
	for _, want := range []string{
		`class="unclaimed-breakdown"`,
		"0.0011 PMT",
		"delegator share",
		"your commission",
		"MsgWithdrawDelegatorReward",
		"MsgWithdrawValidatorCommission",
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("breakdown missing %q:\n%s", want, html)
		}
	}
}

func TestDistributionEscrowReconcileGap(t *testing.T) {
	d := model.Report{
		UnclaimedDelegator:  "0.006854 PMT",
		UnclaimedCommission: "0.000685 PMT",
		ModuleAccounts: []model.ModuleAccountRow{
			{Name: "distribution", Balance: "0.006854 PMT"},
		},
	}
	effect, warn := distributionEscrowReconcile(d)
	if !warn {
		t.Fatal("expected warn when bank matches delegator only but not total")
	}
	if !strings.Contains(effect, "0.000685 PMT") {
		t.Fatalf("expected gap amount in effect, got %q", effect)
	}
}

func TestDistributionEscrowReconcileMatch(t *testing.T) {
	d := model.Report{
		UnclaimedDelegator:  "0.006 PMT",
		UnclaimedCommission: "0.001 PMT",
		ModuleAccounts: []model.ModuleAccountRow{
			{Name: "distribution", Balance: "0.007 PMT"},
		},
	}
	effect, warn := distributionEscrowReconcile(d)
	if warn {
		t.Fatalf("expected no warn when balanced, got %q", effect)
	}
	if !strings.Contains(effect, "matches") {
		t.Fatalf("expected match message, got %q", effect)
	}
}
