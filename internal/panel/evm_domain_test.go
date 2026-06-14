package panel

import (
	"strings"
	"testing"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func TestEVMRPCHealthCards(t *testing.T) {
	d := model.Report{
		EVMRPCOk: true, EVMSynced: true, EVMListening: true,
		EVMBlock: "100", EVMBlockAge: "4.2s", EVMChainID: 290290,
		EVMHTTPEndpoint: "http://localhost:8545", EVMClient: "evmd/v1",
		PendingTx: 2, QueuedTx: 1, EVMPeerCount: 0,
		RPCProbeOK: 8, RPCProbeTotal: 8,
		RPCProbes: []model.RPCProbe{
			{Method: "eth_blockNumber", OK: true, Latency: "12ms"},
			{Method: "eth_chainId", OK: true, Latency: "8ms"},
		},
	}
	out := BuildView(ViewEVM, d)
	for _, want := range []string{
		`eco-domain--rpc-reach`,
		`eco-domain--rpc-head`,
		`eco-domain--rpc-txpool`,
		`eco-domain--rpc-net`,
		`eco-domain--rpc-wallet`,
		`evm-wallet-section`,
		`evm-summary__stack-line">2 pending`,
		`evm-probes-section`,
		`evm-probes__table`,
		`class="dash-subheading">Method probes</h3>`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("EVM RPC view missing %q", want)
		}
	}
	walletIdx := strings.Index(out, `evm-wallet-section`)
	probesIdx := strings.Index(out, `class="dash-subheading">Method probes</h3>`)
	if walletIdx < 0 || probesIdx < 0 || walletIdx > probesIdx {
		t.Fatal("wallet section should appear before method probes")
	}
	for _, absent := range []string{
		`class="dash-subheading">Wallet endpoints</h3>`,
		`MetaMask custom network`,
		`eco-domain--vm`,
		`eco-domain--erc20`,
	} {
		if strings.Contains(out, absent) {
			t.Fatalf("EVM view should not include %q", absent)
		}
	}
}

func TestEVMRPCProbeTableShowsFailure(t *testing.T) {
	d := model.Report{
		EVMRPCOk: true, RPCProbeOK: 1, RPCProbeTotal: 2,
		RPCProbes: []model.RPCProbe{
			{Method: "eth_blockNumber", OK: true, Latency: "5ms"},
			{Method: "eth_syncing", OK: false, Latency: "40ms", Error: "connection refused"},
		},
	}
	out := evmRPCProbeTableHTML(d)
	if !strings.Contains(out, `dash-sources__row--fail`) {
		t.Fatal("failed probe should highlight row")
	}
	if !strings.Contains(out, "connection refused") {
		t.Fatal("failed probe should show error in checks column")
	}
}
