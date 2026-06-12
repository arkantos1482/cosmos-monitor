package panel

import (
	"strings"
	"testing"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func TestSlashingSectionLocalAndNetwork(t *testing.T) {
	d := model.Report{
		BondedCount: 4, JailedCount: 1, TombstonedCount: 0, BelowThreshold: 2,
		BondedPct: 55.5, SlashWindow: "10000", MinSigned: 50,
		UnbondingTime: "21d", MaxValidators: 100, BondDenom: "apmt",
		Validators: []model.Validator{{
			Moniker: "node1", VPFloat: 25, CommissionFloat: 10, Status: "BONDED",
		}},
		Local: model.LocalValidator{
			IsValidator: true, Status: "BONDED", VPPercent: 25, Commission: 10,
			SigningStatus: "ok", Missed: 2,
		},
	}
	chunk := slashingChunk(t, Build(d))
	for _, want := range []string{
		`class="dash-subheading">This validator</h3>`,
		`class="dash-subheading">Network-wide</h3>`,
		`eco-domain--slashing`,
		"signing health",
		`<th>missed</th>`,
		`slashing-summary__health`,
	} {
		if !strings.Contains(chunk, want) {
			t.Fatalf("slashing section missing %q", want)
		}
	}
	for _, gone := range []string{
		`eco-domain--staking`,
		`<th>vp%</th>`,
		"bonded_tokens_pool",
	} {
		if strings.Contains(chunk, gone) {
			t.Fatalf("slashing section should not contain staking content %q", gone)
		}
	}
	slashCardIdx := strings.Index(chunk, `eco-domain--slashing`)
	slashTableIdx := strings.Index(chunk, `<th>missed</th>`)
	if slashCardIdx < 0 || slashTableIdx < 0 || slashCardIdx > slashTableIdx {
		t.Fatal("slashing section should order network card before slashing table")
	}
}

func TestSlashingSectionNoDuplicateFields(t *testing.T) {
	d := model.Report{
		BondedCount: 4, JailedCount: 1, BelowThreshold: 2,
		SlashWindow: "10000", MinSigned: 50,
		Validators: []model.Validator{{
			Moniker: "node1", VPFloat: 25, Status: "BONDED",
		}},
		Local: model.LocalValidator{
			IsValidator: true, Status: "BONDED", SigningStatus: "ok", Missed: 2,
		},
	}
	chunk := slashingChunk(t, Build(d))
	if strings.Count(chunk, "signing health") != 1 {
		t.Fatalf("signing health should appear once in body, got %d", strings.Count(chunk, "signing health"))
	}
	summaryEnd := strings.Index(chunk, `class="dash-subheading">This validator</h3>`)
	if summaryEnd < 0 {
		t.Fatal("expected slashing body")
	}
	summaryCard := chunk[strings.Index(chunk, `class="dash-section__summary-card__body"`):summaryEnd]
	if strings.Contains(summaryCard, "signing health") {
		t.Fatal("signing health belongs in body only, not embedded summary card")
	}
}

func slashingChunk(t *testing.T, out string) string {
	t.Helper()
	idx := strings.Index(out, `class="dash-heading">2. SLASHING</h2>`)
	end := strings.Index(out, `class="dash-heading">3. FEE MARKET</h2>`)
	if idx < 0 || end < 0 {
		t.Fatal("expected slashing and distribution sections")
	}
	return out[idx:end]
}
