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

func TestStakingAccountsTableHTML_delegatorAndOperator(t *testing.T) {
	lv := model.LocalValidator{
		AccountAddr:     "cosmos1akkvh0ahmve830rj4mhkdnqs49kzw23c63nhdx",
		EVMAddr:         "0xEDaCcbBfB7dB3278bc72AeeF66Cc10A96C272a38",
		AccountBalance:  "1.5M PMT",
		OperatorAddr:    "cosmosvaloper1akkvh0ahmve830rj4mhkdnqs49kzw23cl98zp4",
		OperatorBalance: "0.05 PMT",
	}
	out := stakingAccountsTableHTML(lv)
	for _, want := range []string{
		`data-table staking-accounts`,
		`staking-accounts__row--delegator`,
		`staking-accounts__row--operator`,
		`<th>cosmos</th>`,
		`<th>evm</th>`,
		`<th>balance</th>`,
		`cosmos1akkvh0ahmve830rj4mhkdnqs49kzw23c63nhdx`,
		`cosmosvaloper1akkvh0ahmve830rj4mhkdnqs49kzw23cl98zp4`,
		`0xEDaCcbBfB7dB3278bc72AeeF66Cc10A96C272a38`,
		`1.5M PMT`,
		`0.05 PMT`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("staking accounts table missing %q\n%s", want, out)
		}
	}
	for _, gone := range []string{
		`class="id-board"`,
		`<th>bech32</th>`,
		`<th>hex</th>`,
	} {
		if strings.Contains(out, gone) {
			t.Fatalf("staking accounts table should not include %q", gone)
		}
	}
}

func TestStakingAccountsTableHTML_operatorNoEVM(t *testing.T) {
	out := stakingAccountsTableHTML(model.LocalValidator{
		OperatorAddr: "cosmosvaloper1test",
	})
	if strings.Count(out, `<span class="id-empty">—</span>`) < 2 {
		t.Fatalf("operator row should show dashes for empty evm/balance, got:\n%s", out)
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
