package panel

import (
	"strings"
	"testing"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
	"github.com/arkantos1482/cosmos-monitor/internal/report"
)

func TestBuildEVMRPCSection(t *testing.T) {
	d := model.Report{
		EVMRPCOk: true, EVMSynced: true, EVMListening: true,
		EVMBlockAge: "4.2s", EVMBlock: "100", EVMChainID: 290290,
		Network: "pmt", EVMHTTPEndpoint: "http://localhost:8545",
		RPCProbeOK: 9, RPCProbeTotal: 9,
		RPCProbes: []model.RPCProbe{
			{Method: "eth_blockNumber", OK: true, Latency: "12ms"},
		},
		GasPrice: "1 apmt", PendingTx: 2, QueuedTx: 1,
	}
	out := BuildText(d)
	if !strings.Contains(out, "**RPC: OK**") {
		t.Fatal("output should include RPC status line")
	}
	if !strings.Contains(out, "## Probe health") {
		t.Fatal("expected Probe health subsection")
	}
}

func TestFormatProbeExchange(t *testing.T) {
	body := formatProbeExchange(model.RPCProbe{
		Method: "eth_blockNumber", OK: true, Latency: "3ms",
		Request:  `{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}`,
		Response: `{"jsonrpc":"2.0","id":1,"result":"0x10"}`,
	})
	if !strings.Contains(body, "req »") || !strings.Contains(body, "res »") {
		t.Fatalf("expected req/res markers: %q", body)
	}
}

func TestEvmWSEndpoint(t *testing.T) {
	if ws := report.EVMWSEndpoint("http://localhost:8545"); ws != "ws://localhost:8546" {
		t.Fatalf("unexpected ws endpoint: %s", ws)
	}
}
