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
		`class="dash-subheading">Staking</h3>`,
		`class="dash-subheading">Slashing</h3>`,
		`eco-domain--staking`,
		`eco-domain--slashing`,
		"bonded_tokens_pool",
		"unbonding time",
		"signing health",
		`<th>vp%</th>`,
		`<th>missed</th>`,
	} {
		if !strings.Contains(chunk, want) {
			t.Fatalf("staking section missing %q", want)
		}
	}
	stakeIdx := strings.Index(chunk, `eco-domain--staking`)
	slashIdx := strings.Index(chunk, `eco-domain--slashing`)
	slashTableIdx := strings.LastIndex(chunk, `<th>missed</th>`)
	if stakeIdx < 0 || slashIdx < 0 || stakeIdx > slashIdx || slashIdx > slashTableIdx {
		t.Fatal("chain content should be ordered: staking card, slashing card, then slashing table")
	}
}

func TestStakingSectionNoDuplicateFields(t *testing.T) {
	d := model.Report{
		BondedCount: 4, JailedCount: 1, BelowThreshold: 2,
		BondedPct: 55.5, SlashWindow: "10000", MinSigned: 50,
		UnbondingTime: "21d", MaxValidators: 100, BondDenom: "apmt",
		BondedAmt: "10M PMT", NotBonded: "2M PMT", TotalSupply: "12M PMT",
		ModuleAccounts: []model.ModuleAccountRow{
			{Name: "bonded_tokens_pool", Balance: "10M PMT"},
		},
		Validators: []model.Validator{{
			Moniker: "node1", VPFloat: 25, CommissionFloat: 10, Status: "BONDED",
		}},
		Local: model.LocalValidator{
			IsValidator: true, Status: "BONDED", VPPercent: 25, Commission: 10,
			VotingPower: "100", SigningStatus: "ok", Missed: 2,
		},
	}
	chunk := stakingChunk(t, Build(d))
	for _, field := range []string{
		"eco-domain--staking",
		"eco-domain--slashing",
		"bonded_tokens_pool",
		"unbonding time",
		"signing health",
		"val-summary__kpi",
		"val-summary__chip",
	} {
		if n := strings.Count(chunk, field); n != 1 && field != "val-summary__kpi" && field != "val-summary__chip" {
			t.Fatalf("expected %q once in staking section, got %d", field, n)
		}
		if (field == "val-summary__kpi" || field == "val-summary__chip") && strings.Contains(chunk, field) {
			t.Fatalf("staking section should not repeat chain KPI/chip summary alongside tables: %q", field)
		}
	}
	localStake := chunk[strings.Index(chunk, `class="dash-layer__title">This validator</h3>`):strings.Index(chunk, `class="dash-layer__title">Chain</h3>`)]
	if strings.Count(localStake, "25.0%") > 1 {
		t.Fatal("local VP should not appear in both summary and body")
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
