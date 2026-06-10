package panel

import (
	"strings"
	"testing"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func TestIdentityBoardHTML_fourRowsAndFormats(t *testing.T) {
	d := model.Report{
		Moniker: "node1",
		NodeID:  "7c90c68908923b0abf17ce9bb7d79dd405abfe95",
	}
	lv := model.LocalValidator{
		Moniker:         "node1",
		AccountAddr:     "cosmos1akkvh0ahmve830rj4mhkdnqs49kzw23c63nhdx",
		EVMAddr:         "0xEDaCcbBfB7dB3278bc72AeeF66Cc10A96C272a38",
		OperatorAddr:    "cosmosvaloper1akkvh0ahmve830rj4mhkdnqs49kzw23cl98zp4",
		ConsensusAddr:   "31aec3d55f45aa21f7efbcc4257ea9f56c9ad300",
		ConsensusBech32: "cosmosvalcons1w7jy2z0d3m0n0x0x0x0x0x0x0x0x0x0x0x0x0x0x0x0x0x0x",
	}
	// Use empty bech32 to exercise hex→bech32 fallback path in board.
	lv.ConsensusBech32 = ""

	out := identityBoardHTML(d, lv)
	for _, want := range []string{
		`class="id-board"`,
		`id-board__row--account`,
		`id-board__row--operator`,
		`id-board__row--consensus`,
		`id-board__row--p2p`,
		`<span class="id-hrp">cosmos</span>`,
		`<span class="id-hrp">cosmosvaloper</span>`,
		`<span class="id-hrp">cosmosvalcons</span>`,
		`class="id-hex id-hex--evm"`,
		`0xEDaCcbBfB7dB3278bc72AeeF66Cc10A96C272a38`,
		`31AEC3D55F45AA21F7EFBCC4257EA9F56C9AD300`,
		`7c90c68908923b0abf17ce9bb7d79dd405abfe95`,
		`<span class="id-shared">`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("identity board missing %q\n%s", want, out)
		}
	}
	if strings.Contains(out, "operator wallet") || strings.Contains(out, "valoper —") {
		t.Fatal("identity board should not include explanatory captions")
	}
}

func TestIdentityBoardHTML_operatorHexEmpty(t *testing.T) {
	out := identityBoardHTML(model.Report{}, model.LocalValidator{
		OperatorAddr: "cosmosvaloper1test",
	})
	if strings.Count(out, `<span class="id-empty">—</span>`) < 2 {
		t.Fatalf("operator and p2p hex/bech32 empties expected, got:\n%s", out)
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
