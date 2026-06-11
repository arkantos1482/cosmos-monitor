package panel

import (
	"strings"
	"testing"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func TestValidatorsNoDomainCards(t *testing.T) {
	d := model.Report{
		BondedCount: 4, JailedCount: 1, TombstonedCount: 0, BelowThreshold: 2,
		BondedPct: 55.5, SlashWindow: "10000", MinSigned: 50,
		UnbondingTime: "21d", MaxValidators: 100,
		ModuleAccounts: []model.ModuleAccountRow{
			{Name: "bonded_tokens_pool", Balance: "10M PMT", Address: "cosmos1bonded"},
		},
	}
	chunk := validatorsChunk(t, Build(d))
	for _, want := range []string{
		`class="val-summary"`,
		`class="dash-subheading">Stake</h3>`,
		`class="dash-subheading">Slashing</h3>`,
		"below min signed",
	} {
		if !strings.Contains(chunk, want) {
			t.Fatalf("validators missing %q", want)
		}
	}
	for _, gone := range []string{
		`eco-domain--staking`,
		`eco-domain--slashing`,
		"bonded_tokens_pool",
		"unbonding time",
		`eco-domain__divider">Governance params`,
	} {
		if strings.Contains(chunk, gone) {
			t.Fatalf("validators should not duplicate economics chain modules: %q", gone)
		}
	}
}
