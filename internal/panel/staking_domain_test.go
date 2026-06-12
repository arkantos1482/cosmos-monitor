package panel

import (
	"strings"
	"testing"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func TestStakingSectionConsolidatesLocalAndChain(t *testing.T) {
	d := model.Report{
		BondedCount: 4, JailedCount: 1, TombstonedCount: 0, BelowThreshold: 2,
		BondedPct: 55.5, SlashWindow: "10000", MinSigned: 50,
		UnbondingTime: "21d", MaxValidators: 100, BondDenom: "apmt",
		ModuleAccounts: []model.ModuleAccountRow{
			{Name: "bonded_tokens_pool", Balance: "10M PMT", Address: "cosmos1bonded"},
		},
		Validators: []model.Validator{{
			Moniker: "node1", VPFloat: 25, CommissionFloat: 10, Status: "BONDED",
		}},
		Local: model.LocalValidator{
			IsValidator: true, Status: "BONDED", VPPercent: 25, Commission: 10,
			SigningStatus: "ok", Missed: 2,
		},
	}
	chunk := stakingChunk(t, Build(d))
	for _, want := range []string{
		`class="dash-layer__title">This validator</h3>`,
		`class="dash-layer__title">Chain</h3>`,
		`class="dash-subheading">Stake</h3>`,
		`class="dash-subheading">Signing</h3>`,
		`class="dash-subheading">Validator stake</h3>`,
		`class="dash-subheading">Validator slashing</h3>`,
		`eco-domain--staking`,
		`eco-domain--slashing`,
		"bonded_tokens_pool",
		"unbonding time",
		"signing health",
		`class="val-summary"`,
	} {
		if !strings.Contains(chunk, want) {
			t.Fatalf("staking section missing %q", want)
		}
	}
}

func stakingChunk(t *testing.T, out string) string {
	t.Helper()
	idx := strings.Index(out, `class="dash-heading">1. STAKING</h2>`)
	end := strings.Index(out, `class="dash-heading">2. VALIDATOR SET</h2>`)
	if idx < 0 || end < 0 {
		t.Fatal("expected staking and validator set sections")
	}
	return out[idx:end]
}
