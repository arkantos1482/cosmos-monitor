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
		EVMDenomName: "PMT", EVMDenomSymbol: "PMT", EVMDenomDecimals: 18,
		PendingTx: 2, QueuedTx: 1, EVMPeerCount: 0,
		RPCProbeOK: 10, RPCProbeTotal: 10,
		RPCProbes: []model.RPCProbe{
			{Method: "eth_blockNumber", OK: true, Latency: "12ms"},
			{Method: "eth_chainId", OK: true, Latency: "8ms"},
			{Method: "eth_chainId", Transport: "ws", OK: true, Latency: "9ms"},
		},
	}
	out := BuildView(ViewEVM, d)
	for _, want := range []string{
		`class="dash-subheading">MetaMask</h3>`,
		`eco-domain--rpc-metamask`,
		`eco-domain--rpc-reach`,
		`eco-domain--rpc-head`,
		`eco-domain--rpc-txpool`,
		`eco-domain--rpc-net`,
		`network name`,
		`currency symbol`,
		`decimals`,
		`HTTP probes`,
		`WS probes`,
		`evm-probes__group-title">HTTP`,
		`evm-probes__group-title">WebSocket`,
		`evm-summary__hero-label">HTTP probes`,
		`evm-summary__hero-label">WS probes`,
		`evm-summary__stack-line">2 pending`,
		`evm-probes-section`,
		`evm-probes__table`,
		`class="dash-subheading">Method probes</h3>`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("EVM RPC view missing %q", want)
		}
	}
	metaIdx := strings.Index(out, `class="dash-subheading">MetaMask</h3>`)
	reachIdx := strings.Index(out, `eco-domain--rpc-reach`)
	probesIdx := strings.Index(out, `class="dash-subheading">Method probes</h3>`)
	if metaIdx < 0 || reachIdx < 0 || probesIdx < 0 || !(metaIdx < reachIdx && reachIdx < probesIdx) {
		t.Fatal("MetaMask subsection should appear before health cards and method probes")
	}
	for _, absent := range []string{
		`class="dash-subheading">Wallet endpoints</h3>`,
		`eco-domain--rpc-wallet`,
		`evm-wallet-section`,
		`eco-domain__divider">MetaMask custom network`,
		`HTTP endpoint`,
		`Network name: PMT`,
		`wallet endpoints`,
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

func TestEVMWalletLabelsFromBankMetadata(t *testing.T) {
	d := model.Report{
		EVMDenomName: "Acme Chain", EVMDenomSymbol: "ACM", EVMDenomDecimals: 6,
	}
	if got := evmNetworkName(d); got != "Acme Chain" {
		t.Fatalf("network name = %q, want Acme Chain", got)
	}
	if got := evmCurrencySymbol(d); got != "ACM" {
		t.Fatalf("currency symbol = %q, want ACM", got)
	}
	if got := evmCurrencyDecimals(d); got != "6" {
		t.Fatalf("decimals = %q, want 6", got)
	}
}

func TestEVMWalletLabelsFallbackWithoutMetadata(t *testing.T) {
	d := model.Report{Network: "pmt", EVMDenom: "apmt"}
	if got := evmNetworkName(d); got != "PMT" {
		t.Fatalf("network name fallback = %q, want PMT", got)
	}
	if got := evmCurrencySymbol(d); got != "PMT" {
		t.Fatalf("currency symbol fallback = %q, want PMT", got)
	}
}
