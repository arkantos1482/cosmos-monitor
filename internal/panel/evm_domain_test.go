package panel

import (
	"strings"
	"testing"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func TestEVMDomainCards(t *testing.T) {
	d := model.Report{
		EVMRPCOk: true, EVMSynced: true, EVMBlock: "100", EVMChainID: 290290,
		EVMDenom: "apmt", Precompiles: []string{"0x01", "0x09"},
		HistoryWindow: "8192", HardforkLondon: "0", HardforkShanghai: "100",
		ERC20Enabled: true, TokenPairs: []model.TokenPair{{Enabled: true}},
		Local: model.LocalValidator{EVMAddr: "0xLOCAL"},
		RPCProbeOK: 1, RPCProbeTotal: 1,
		RPCProbes: []model.RPCProbe{{Method: "eth_blockNumber", OK: true, Latency: "1ms"}},
	}
	out := BuildView(ViewEVM, d)
	for _, want := range []string{
		`eco-domain--vm`,
		`eco-domain--erc20`,
		"precompiles",
		"history window",
		"shanghai_block",
		"enable_erc20",
		"0xLOCAL",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("EVM view missing %q", want)
		}
	}
}
