package panel

import (
	"strings"
	"testing"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func TestNodePageLocalStateOnly(t *testing.T) {
	d := model.Report{
		SlashWindow: "10000", MinSigned: 50, CommunityTax: "2%",
		UnbondingTime: "21d", MaxValidators: 100, BondDenom: "apmt",
		Local: model.LocalValidator{
			IsValidator: true, Status: "BONDED", VPPercent: 25, Commission: 10,
			Outstanding: "100 PMT", CommissionEarned: "5 PMT", Missed: 2,
			SigningStatus: "ok", EVMAddr: "0xvalidator",
		},
	}
	out := BuildView(ViewNode, d)
	for _, want := range []string{
		`class="dash-subheading">Staking</h3>`,
		`class="dash-subheading">Rewards</h3>`,
		`class="dash-subheading">Slashing</h3>`,
		"outstanding rewards",
		"commission earned",
		"signing health",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("node view missing local state %q", want)
		}
	}
	for _, gone := range []string{
		`eco-domain--staking`,
		`eco-domain--slashing`,
		`eco-domain--distribution`,
		"unbonding time",
		"max validators",
		"community tax",
		"bond denom",
		`eco-domain__divider">Governance params`,
	} {
		if strings.Contains(out, gone) {
			t.Fatalf("node view should not show chain-level %q", gone)
		}
	}
}
