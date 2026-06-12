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
		`class="dash-layer__title">Staking</h3>`,
		`class="dash-layer__title">Slashing</h3>`,
		`class="dash-subheading">This validator</h3>`,
		`class="dash-subheading">Chain</h3>`,
		`eco-domain--staking`,
		`eco-domain--slashing`,
		"bonded_tokens_pool",
		"unbonding time",
		"signing health",
		`<th>vp%</th>`,
		`<th>missed</th>`,
		`staking-summary__vp`,
	} {
		if !strings.Contains(chunk, want) {
			t.Fatalf("staking section missing %q", want)
		}
	}
	stakeLayer := chunk[strings.Index(chunk, `class="dash-layer__title">Staking</h3>`):strings.Index(chunk, `class="dash-layer__title">Slashing</h3>`)]
	slashLayer := chunk[strings.Index(chunk, `class="dash-layer__title">Slashing</h3>`):]
	stakeCardIdx := strings.Index(stakeLayer, `eco-domain--staking`)
	stakeTableIdx := strings.Index(stakeLayer, `<th>vp%</th>`)
	slashCardIdx := strings.Index(slashLayer, `eco-domain--slashing`)
	slashTableIdx := strings.Index(slashLayer, `<th>missed</th>`)
	if stakeCardIdx < 0 || stakeTableIdx < 0 || stakeCardIdx > stakeTableIdx {
		t.Fatal("staking layer should order chain card before stake table")
	}
	if slashCardIdx < 0 || slashTableIdx < 0 || slashCardIdx > slashTableIdx {
		t.Fatal("slashing layer should order chain card before slashing table")
	}
	stakeLayerIdx := strings.Index(chunk, `class="dash-layer__title">Staking</h3>`)
	slashLayerIdx := strings.Index(chunk, `class="dash-layer__title">Slashing</h3>`)
	if stakeLayerIdx < 0 || slashLayerIdx < 0 || stakeLayerIdx > slashLayerIdx {
		t.Fatal("staking layer should precede slashing layer")
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
	} {
		if n := strings.Count(chunk, field); n != 1 {
			t.Fatalf("expected %q once in staking section, got %d", field, n)
		}
	}
	for _, field := range []string{"val-summary__kpi", "val-summary__chip"} {
		if strings.Contains(chunk, field) {
			t.Fatalf("staking section should not repeat chain KPI/chip summary alongside tables: %q", field)
		}
	}
	summaryEnd := strings.Index(chunk, `class="dash-layer__title">Staking</h3>`)
	if summaryEnd < 0 {
		t.Fatal("expected staking layer")
	}
	summary := chunk[:summaryEnd]
	body := chunk[summaryEnd:]
	if strings.Count(summary, "25.0%") != 1 {
		t.Fatalf("node VP should appear once in embedded summary, got %d", strings.Count(summary, "25.0%"))
	}
	if strings.Contains(summary, "signing health") {
		t.Fatal("node signing health belongs in body only, not embedded summary")
	}
	if strings.Count(body, "signing health") != 1 {
		t.Fatalf("node signing health should appear once in body, got %d", strings.Count(body, "signing health"))
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
