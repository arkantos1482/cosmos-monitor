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
		RPCProbeOK: 8, RPCProbeTotal: 8,
		RPCProbes: []model.RPCProbe{
			{Method: "eth_blockNumber", OK: true, Latency: "12ms"},
		},
		PendingTx: 2, QueuedTx: 1,
	}
	out := Build(d)
	if strings.Contains(out, "<strong>RPC: OK</strong>") {
		t.Fatal("output should not use legacy StrongLine RPC callout")
	}
	if !strings.Contains(out, `class="evm-summary"`) {
		t.Fatal("output should include EVM summary")
	}
	if !strings.Contains(out, `evm-summary__probe`) {
		t.Fatal("EVM summary should include probe dots")
	}
	if !strings.Contains(out, `class="dash-subheading">Probe health</h3>`) {
		t.Fatal("expected Probe health subsection")
	}
}

func TestFormatProbeExchange(t *testing.T) {
	body := formatProbeExchangeHTML(model.RPCProbe{
		Method: "eth_blockNumber", OK: true, Latency: "3ms",
		Request:  `{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}`,
		Response: `{"jsonrpc":"2.0","id":1,"result":"0x10"}`,
	})
	for _, want := range []string{`dash-sources__tag">req`, `dash-sources__tag">res`, `json-key`, `json-block`} {
		if !strings.Contains(body, want) {
			t.Fatalf("expected %q in probe exchange HTML: %q", want, body)
		}
	}
}

func TestEvmWSEndpoint(t *testing.T) {
	if ws := report.EVMWSEndpoint("http://localhost:8545"); ws != "ws://localhost:8546" {
		t.Fatalf("unexpected ws endpoint: %s", ws)
	}
}
