package panel

import (
	"strings"
	"testing"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func TestStakingSectionLocalAndNetwork(t *testing.T) {
	d := model.Report{
		BondedCount: 4, JailedCount: 1, BelowThreshold: 2,
		BondedPct: 55.5, SlashWindow: "10000", MinSigned: 50,
		UnbondingTime: "21d", MaxValidators: 100, BondDenom: "apmt",
		BondedAmt: "10M PMT",
		ModuleAccounts: []model.ModuleAccountRow{
			{Name: "bonded_tokens_pool", Balance: "10M PMT", Address: "cosmos1bonded"},
		},
		Validators: []model.Validator{{
			Moniker: "node1", Operator: "cosmosvaloper1abc",
			VPFloat: 25, CommissionFloat: 10, Status: "BONDED", IsLocal: true,
		}},
		Local: model.LocalValidator{
			IsValidator: true, Moniker: "node1", Status: "bonded", VPPercent: 25, Commission: 10,
			VotingPower: "100 PMT", SigningStatus: "ok", Missed: 2,
			AccountAddr: "cosmos1account", EVMAddr: "0xACCOUNT",
			OperatorAddr: "cosmosvaloper1abc", CommissionEarned: "0.1 PMT",
			Outstanding: "0.5 PMT", LiquidBalance: "1M PMT", DelegatorCount: 1,
			Delegations: []model.DelegationRow{{
				Delegator: "cosmos1delegator", EVMAddr: "0xDELEGATOR",
				Balance: "100 PMT", LiquidBalance: "50 PMT", Shares: "100000000000000000000",
				IsLocal: true,
			}},
		},
	}
	chunk := stakingChunk(t, Build(d))
	for _, want := range []string{
		`class="dash-subheading">This validator</h3>`,
		`class="dash-subheading">Network-wide</h3>`,
		`eco-domain--staking`,
		"bonded_tokens_pool",
		"unbonding time",
		`<th>operator</th>`,
		`<th class="data-table__num">vp%</th>`,
		`<th class="data-table__num">commission</th>`,
		`<th class="data-table__center">status</th>`,
		`<code>cosmosvaloper1abc</code>`,
		`class="data-table__row--local" title="this node"`,
		`staking-summary__kpi`,
		`staking-summary__kpis--network`,
		`data-table--delegations`,
		`<th>delegator</th>`,
		`<th class="data-table__num">delegated</th>`,
		`<th class="data-table__num">spendable</th>`,
		`id-addr-table`,
		`<td class="id-addr-table__role">operator</td>`,
		`<td class="id-addr-table__role">account</td>`,
		`<td class="id-addr-table__role">evm</td>`,
		`class="metric-split"`,
		`class="id-dual"`,
		`0xDELEGATOR`,
		`<th class="data-table__num">delegation shares</th>`,
		`liquid balance`,
		`delegators`,
		`data-table--staking-set`,
		`100 PMT`,
	} {
		if !strings.Contains(chunk, want) {
			t.Fatalf("staking section missing %q", want)
		}
	}
	for _, gone := range []string{
		`eco-domain--slashing`,
		`<th>missed</th>`,
		"signing health",
		`Next proposer`,
		`val-summary__proposer`,
		`class="id-board"`,
		`outstanding rewards`,
		`commission earned`,
		`100000000000000000000`,
	} {
		if strings.Contains(chunk, gone) {
			t.Fatalf("staking section should not contain %q", gone)
		}
	}
	stakeCardIdx := strings.Index(chunk, `eco-domain--staking`)
	stakeTableIdx := strings.Index(chunk, `<th class="data-table__num">vp%</th>`)
	if stakeCardIdx < 0 || stakeTableIdx < 0 || stakeCardIdx > stakeTableIdx {
		t.Fatal("staking section should order network card before stake table")
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
			Moniker: "node1", Operator: "cosmosvaloper1abc",
			VPFloat: 25, CommissionFloat: 10, Status: "BONDED", IsLocal: true,
		}},
		Local: model.LocalValidator{
			IsValidator: true, Status: "BONDED", VPPercent: 25, Commission: 10,
			VotingPower: "100", SigningStatus: "ok", Missed: 2,
		},
	}
	chunk := stakingChunk(t, Build(d))
	for _, field := range []string{
		"eco-domain--staking",
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
	summaryEnd := strings.Index(chunk, `class="dash-subheading">This validator</h3>`)
	if summaryEnd < 0 {
		t.Fatal("expected staking body")
	}
	summary := chunk[:summaryEnd]
	if strings.Count(summary, "25.0%") != 1 {
		t.Fatalf("node VP should appear once in embedded summary, got %d", strings.Count(summary, "25.0%"))
	}
}

func stakingChunk(t *testing.T, out string) string {
	t.Helper()
	idx := strings.Index(out, `class="dash-heading">1. STAKING</h2>`)
	end := strings.Index(out, `class="dash-heading">2. SLASHING</h2>`)
	if idx < 0 || end < 0 {
		t.Fatal("expected staking and slashing sections")
	}
	return out[idx:end]
}
