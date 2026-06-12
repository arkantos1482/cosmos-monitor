package panel

import (
	"strings"
	"testing"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func TestValidatorIdentityBoardHTML_consensusAndP2P(t *testing.T) {
	d := model.Report{
		Moniker: "node1",
		NodeID:  "7c90c68908923b0abf17ce9bb7d79dd405abfe95",
	}
	lv := model.LocalValidator{
		Moniker:       "node1",
		ConsensusAddr: "31aec3d55f45aa21f7efbcc4257ea9f56c9ad300",
	}
	out := validatorIdentityBoardHTML(d, lv)
	for _, want := range []string{
		`class="id-board"`,
		`id-board__row--consensus`,
		`id-board__row--p2p`,
		`<span class="id-hrp">cosmosvalcons</span>`,
		`31AEC3D55F45AA21F7EFBCC4257EA9F56C9AD300`,
		`7c90c68908923b0abf17ce9bb7d79dd405abfe95`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("validator identity board missing %q\n%s", want, out)
		}
	}
	for _, gone := range []string{
		`id-board__row--account`,
		`id-board__row--operator`,
	} {
		if strings.Contains(out, gone) {
			t.Fatalf("validator identity board should not include %q", gone)
		}
	}
}

func TestStakingDelegatorsTable(t *testing.T) {
	d := model.Report{
		Local: model.LocalValidator{
			IsValidator:      true,
			OperatorAddr:     "cosmosvaloper1akkvh0ahmve830rj4mhkdnqs49kzw23cl98zp4",
			CommissionEarned: "0.05 PMT",
			Delegations: []model.DelegationRow{{
				Delegator: "cosmos1akkvh0ahmve830rj4mhkdnqs49kzw23c63nhdx",
				EVMAddr:   "0xEDaCcbBfB7dB3278bc72AeeF66Cc10A96C272a38",
				Balance:   "100M PMT", LiquidBalance: "10M PMT", Shares: "100000000000000000000",
				IsLocal:   true,
			}, {
				Delegator: "cosmos1otherdelegator",
				EVMAddr:   "0xOTHER",
				Balance:   "5M PMT", LiquidBalance: "1M PMT",
			}},
		},
	}
	chunk := stakingChunk(t, Build(d))
	for _, want := range []string{
		`<th>address</th>`,
		`<th class="data-table__num">delegated</th>`,
		`<th class="data-table__num">liquid</th>`,
		`class="id-dual"`,
		`data-table--delegations`,
		`table-scroll--fit`,
		`akkvh0ahmve830rj4mhkdnqs49kzw23c63nhdx`,
		`0xEDaCcbBfB7dB3278bc72AeeF66Cc10A96C272a38`,
		`otherdelegator`,
		`100M PMT`,
		`10M PMT`,
		`5M PMT`,
		`class="data-table__row--local" title="this node"`,
	} {
		if !strings.Contains(chunk, want) {
			t.Fatalf("staking delegators missing %q", want)
		}
	}
	for _, gone := range []string{
		`validator_account`,
		`data-table staking-accounts`,
		`<th>evm</th>`,
		`<th>delegator</th>`,
	} {
		if strings.Contains(chunk, gone) {
			t.Fatalf("staking section should not include %q", gone)
		}
	}
}

func TestLongestCommonPrefix(t *testing.T) {
	a := "akkvh0ahmve830rj4mhkdnqs49kzw23c63nhdx"
	b := "akkvh0ahmve830rj4mhkdnqs49kzw23cl98zp4"
	got := longestCommonPrefix(a, b)
	want := "akkvh0ahmve830rj4mhkdnqs49kzw23c"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
