package panel

import (
	"fmt"
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
		RPCProbeOK: 10, RPCProbeTotal: 10,
		RPCProbes: []model.RPCProbe{
			{Method: "eth_blockNumber", OK: true, Latency: "12ms"},
			{Method: "eth_chainId", Transport: "ws", OK: true, Latency: "11ms"},
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
	if !strings.Contains(out, `class="dash-subheading">Method probes</h3>`) {
		t.Fatal("EVM section should render inline method probe table")
	}
	if strings.Contains(out, "`x/vm`") {
		t.Fatal("EVM section intro should not reference x/vm module")
	}
}

func TestEVMDataSourcesProvenance(t *testing.T) {
	httpMethods := []string{
		"eth_blockNumber", "eth_chainId", "eth_syncing", "txpool_status",
		"net_peerCount", "net_listening", "web3_clientVersion", "eth_getBlockByNumber",
	}
	wsMethods := []string{"eth_chainId", "net_version"}
	var probes []model.RPCProbe
	for _, method := range httpMethods {
		probes = append(probes, model.RPCProbe{
			Method: method, OK: true, Latency: "3ms",
			Request:  fmt.Sprintf(`{"jsonrpc":"2.0","method":"%s","params":[],"id":1}`, method),
			Response: `{"jsonrpc":"2.0","id":1,"result":"0x1"}`,
		})
	}
	for _, method := range wsMethods {
		probes = append(probes, model.RPCProbe{
			Method: method, Transport: "ws", OK: true, Latency: "3ms",
			Request:  fmt.Sprintf(`{"jsonrpc":"2.0","method":"%s","params":[],"id":1}`, method),
			Response: `{"jsonrpc":"2.0","id":1,"result":"0x1"}`,
		})
	}
	d := model.Report{
		EVMRPCOk: true, EVMSynced: true, EVMBlock: "100", EVMChainID: 290290,
		EVMHTTPEndpoint: "http://localhost:8545",
		EVMWSEndpoint:   "ws://localhost:8546",
		RPCProbes:       probes,
		Exchanges: []model.SourceExchange{
			{
				Kind: "http", Method: "GET",
				URL:      "http://localhost:1317/cosmos/evm/vm/v1/params",
				Request:  "(none)",
				Response: `{"params":{}}`,
				OK:       true, Latency: "2ms",
			},
			{
				Kind: "http", Method: "GET",
				URL:      "http://localhost:1317/cosmos/bank/v1beta1/denoms_metadata/apmt",
				Request:  "(none)",
				Response: `{"metadata":{"name":"PMT","symbol":"PMT"}}`,
				OK:       true, Latency: "2ms",
			},
		},
	}
	out := BuildViewWithOptions(ViewEVM, d, Options{ShowSources: true})
	for _, want := range []string{
		`class="dash-sources"`,
		`>Data sources</summary>`,
		`dash-sources__exchange`,
		`dash-sources__tag">req`,
		`dash-sources__tag">res`,
		`/cosmos/evm/vm/v1/params`,
		`/cosmos/bank/v1beta1/denoms_metadata/apmt`,
		`WS net_version`,
		`WS eth_chainId`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("EVM data sources missing %q", want)
		}
	}
	for _, method := range httpMethods {
		if !strings.Contains(out, "POST "+method) {
			t.Fatalf("EVM data sources missing probe %q", method)
		}
	}
	if strings.Contains(out, `/cosmos/evm/feemarket/`) {
		t.Fatal("EVM data sources should not include unrelated REST traces")
	}
	probesIdx := strings.Index(out, `class="dash-subheading">Method probes</h3>`)
	sourcesIdx := strings.Index(out, `class="dash-sources"`)
	if probesIdx < 0 || sourcesIdx < 0 || sourcesIdx < probesIdx {
		t.Fatal("data sources should render after probe table")
	}
	outDefault := BuildView(ViewEVM, d)
	if strings.Contains(outDefault, `class="dash-sources"`) {
		t.Fatal("EVM section should hide data sources by default")
	}
}

func TestEvmWSEndpoint(t *testing.T) {
	if ws := report.EVMWSEndpoint("http://localhost:8545"); ws != "ws://localhost:8546" {
		t.Fatalf("unexpected ws endpoint: %s", ws)
	}
}

func TestEVMSummaryLayout(t *testing.T) {
	d := model.Report{
		EVMRPCOk: true, EVMSynced: true, EVMListening: true,
		EVMBlock: "12345", EVMBlockAge: "3.1s", EVMChainID: 290290,
		EVMHTTPEndpoint: "http://localhost:8545",
		RPCProbeOK: 10, RPCProbeTotal: 10,
		RPCProbes: []model.RPCProbe{
			{Method: "eth_blockNumber", OK: true, Latency: "10ms"},
			{Method: "eth_chainId", OK: true, Latency: "14ms"},
			{Method: "net_version", Transport: "ws", OK: true, Latency: "9ms"},
		},
		PendingTx: 2, QueuedTx: 1,
	}
	out := BuildView(ViewEVM, d)
	for _, want := range []string{
		`evm-summary__hero`,
		`evm-summary__probe-heroes`,
		`evm-summary__hero-label">HTTP probes`,
		`evm-summary__hero-label">WS probes`,
		`evm-summary__kpis`,
		`evm-summary__kpi-label">block age`,
		`evm-summary__stack-line">2 pending`,
		`evm-summary__stack-line">1 queued`,
		`probe pass rate`,
		`11ms avg`,
		`evm-summary__badges`,
		`badge--ok">RPC OK`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("EVM summary missing %q", want)
		}
	}
	for _, gone := range []string{
		`evm-summary__meta`,
		`evm-summary__probes-label`,
		`evm-summary__detail`,
	} {
		if strings.Contains(out, gone) {
			t.Fatalf("EVM summary should not include legacy %q", gone)
		}
	}
}
