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

func TestStakingIdentityBoardHTML_accountAndOperator(t *testing.T) {
	lv := model.LocalValidator{
		Moniker:      "node1",
		AccountAddr:  "cosmos1akkvh0ahmve830rj4mhkdnqs49kzw23c63nhdx",
		EVMAddr:      "0xEDaCcbBfB7dB3278bc72AeeF66Cc10A96C272a38",
		OperatorAddr: "cosmosvaloper1akkvh0ahmve830rj4mhkdnqs49kzw23cl98zp4",
	}
	out := stakingIdentityBoardHTML(model.Report{Moniker: "node1"}, lv)
	for _, want := range []string{
		`class="id-board"`,
		`id-board__row--account`,
		`id-board__row--operator`,
		`<span class="id-hrp">cosmos</span>`,
		`<span class="id-hrp">cosmosvaloper</span>`,
		`class="id-hex id-hex--evm"`,
		`0xEDaCcbBfB7dB3278bc72AeeF66Cc10A96C272a38`,
		`<span class="id-shared">`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("staking identity board missing %q\n%s", want, out)
		}
	}
	for _, gone := range []string{
		`id-board__row--consensus`,
		`id-board__row--p2p`,
	} {
		if strings.Contains(out, gone) {
			t.Fatalf("staking identity board should not include %q", gone)
		}
	}
}

func TestStakingIdentityBoardHTML_operatorHexEmpty(t *testing.T) {
	out := stakingIdentityBoardHTML(model.Report{}, model.LocalValidator{
		OperatorAddr: "cosmosvaloper1test",
	})
	if strings.Count(out, `<span class="id-empty">—</span>`) < 1 {
		t.Fatalf("operator hex empty expected, got:\n%s", out)
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
