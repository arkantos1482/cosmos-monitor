package panel

import (
	"strings"
	"testing"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func TestUnclaimedStackSummaryOmitsClaimHints(t *testing.T) {
	html := unclaimedStackHTML(unclaimedStackFromLocal(model.LocalValidator{
		Outstanding:      "0.001 PMT",
		CommissionEarned: "0.0001 PMT",
	}))
	if strings.Contains(html, "MsgWithdraw") {
		t.Fatalf("summary stack should not include claim hints:\n%s", html)
	}
}

func TestLocalUnclaimedBreakdownHTML(t *testing.T) {
	html := localUnclaimedBreakdownHTML(model.LocalValidator{
		Outstanding:      "0.001 PMT",
		CommissionEarned: "0.0001 PMT",
	})
	for _, want := range []string{
		`class="unclaimed-stack"`,
		`unclaimed-stack__head-label">unclaimed total`,
		"0.0011 PMT",
		"delegator share",
		"your commission",
		"unclaimed-stack__op",
		"MsgWithdrawDelegatorReward",
		"MsgWithdrawValidatorCommission",
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("breakdown missing %q:\n%s", want, html)
		}
	}
}

