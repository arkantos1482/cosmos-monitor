package panel

import (
	"strings"
	"testing"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func sampleExchanges() []model.SourceExchange {
	return []model.SourceExchange{
		{
			Kind: "http", Method: "GET",
			URL:      "http://localhost:26657/status",
			Request:  "(none)",
			Response: `{"result":{"node_info":{"moniker":"node1"}}}`,
			OK:       true, Latency: "2ms",
		},
		{
			Kind: "http", Method: "GET",
			URL:      "http://localhost:1317/cosmos/distribution/v1beta1/params",
			Request:  "(none)",
			Response: `{"params":{"community_tax":"0.02"}}`,
			OK:       true, Latency: "3ms",
		},
	}
}

func TestSourceExchangesHTML(t *testing.T) {
	html := sourceExchangesHTML(sampleExchanges())
	for _, want := range []string{
		`dash-sources__table`,
		`dash-sources__verb">GET`,
		`dash-sources__path">/status`,
		`dash-sources__exchange`,
		`dash-sources__tag">req`,
		`dash-sources__tag">res`,
		`json-block`,
		`json-key`,
		`distribution/v1beta1/params`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("missing %q in:\n%s", want, html)
		}
	}
}

func TestSourceExchangeTableWrapsLongPath(t *testing.T) {
	long := "/cosmos/distribution/v1beta1/validators/cosmosvaloper15hr4x4rfj0y82puk74xegugn5s5clphzcfej3e/outstanding_rewards"
	html := renderSourceExchangeTable([]model.SourceExchange{{
		Kind: "http", Method: "GET",
		URL: "http://localhost:1317" + long,
		OK:  true, Latency: "2ms",
	}})
	for _, want := range []string{
		`dash-sources__verb">GET`,
		`dash-sources__path">` + long,
		`dash-sources__status">ok`,
		`dash-sources__latency">2ms`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("missing %q in:\n%s", want, html)
		}
	}
}

func TestExchangesForViewNode(t *testing.T) {
	all := sampleExchanges()
	got := exchangesForView(ViewNode, all)
	if len(got) != 1 || !strings.Contains(got[0].URL, "/status") {
		t.Fatalf("node view should only match comet status, got %d", len(got))
	}
}

func TestExchangesForViewDistribution(t *testing.T) {
	all := append(sampleExchanges(), []model.SourceExchange{
		{
			Kind: "http", Method: "GET",
			URL:      "http://localhost:1317/cosmos/staking/v1beta1/validators?status=BOND_STATUS_BONDED",
			Request:  "(none)",
			Response: `{"validators":[]}`,
			OK:       true, Latency: "4ms",
		},
		{
			Kind: "http", Method: "GET",
			URL:      "http://localhost:1317/cosmos/distribution/v1beta1/community_pool",
			Request:  "(none)",
			Response: `{"pool":[]}`,
			OK:       true, Latency: "5ms",
		},
		{
			Kind: "http", Method: "GET",
			URL:      "http://localhost:1317/cosmos/mint/v1beta1/inflation",
			Request:  "(none)",
			Response: `{"inflation":"0"}`,
			OK:       true, Latency: "6ms",
		},
	}...)
	got := sourcesForView(ViewDistribution, model.Report{Exchanges: all})
	if len(got) != 4 {
		t.Fatalf("distribution view should match status + distribution + staking endpoints, got %d", len(got))
	}
	for _, want := range []string{"/status", "distribution", "staking"} {
		found := false
		for _, e := range got {
			if strings.Contains(e.URL, want) {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("distribution sources missing %q", want)
		}
	}
}

func TestExchangesForViewHome(t *testing.T) {
	all := sampleExchanges()
	got := exchangesForView(ViewHome, all)
	if len(got) != len(all) {
		t.Fatalf("home view should include all exchanges, got %d want %d", len(got), len(all))
	}
}

func TestExchangesForViewEVM(t *testing.T) {
	all := append(sampleExchanges(), []model.SourceExchange{
		{
			Kind: "jsonrpc", Method: "POST",
			URL:      "http://localhost:8545",
			Request:  `{"jsonrpc":"2.0","method":"eth_chainId","params":[],"id":1}`,
			Response: `{"jsonrpc":"2.0","id":1,"result":"0x46f32"}`,
			OK:       true, Latency: "4ms",
		},
		{
			Kind: "http", Method: "GET",
			URL:      "http://localhost:1317/cosmos/evm/vm/v1/params",
			Request:  "(none)",
			Response: `{"params":{}}`,
			OK:       true, Latency: "2ms",
		},
	}...)
	got := exchangesForView(ViewEVM, all)
	if len(got) != 2 {
		t.Fatalf("EVM view should match jsonrpc and vm params, got %d", len(got))
	}
}

func TestEVMSourcesIncludeAllRPCProbes(t *testing.T) {
	d := model.Report{
		EVMHTTPEndpoint: "http://localhost:8545",
		EVMWSEndpoint:   "ws://localhost:8546",
		RPCProbes: []model.RPCProbe{
			{Method: "eth_blockNumber", OK: true, Latency: "3ms",
				Request: `{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}`,
				Response: `{"jsonrpc":"2.0","id":1,"result":"0x10"}`},
			{Method: "eth_syncing", OK: true, Latency: "4ms",
				Request: `{"jsonrpc":"2.0","method":"eth_syncing","params":[],"id":1}`,
				Response: `{"jsonrpc":"2.0","id":1,"result":false}`},
			{Method: "net_version", Transport: "ws", OK: true, Latency: "5ms",
				Request: `{"jsonrpc":"2.0","method":"net_version","params":[],"id":1}`,
				Response: `{"jsonrpc":"2.0","id":1,"result":"290290"}`},
		},
		Exchanges: []model.SourceExchange{
			{
				Kind: "http", Method: "GET",
				URL:      "http://localhost:1317/cosmos/evm/vm/v1/params",
				Request:  "(none)",
				Response: `{"params":{}}`,
				OK:       true, Latency: "2ms",
			},
		},
	}
	got := sourcesForView(ViewEVM, d)
	if len(got) != 4 {
		t.Fatalf("EVM sources want 4 entries (1 REST + 3 probes), got %d", len(got))
	}
	for _, want := range []string{"eth_blockNumber", "eth_syncing", "net_version", "/cosmos/evm/vm/v1/params"} {
		found := false
		for _, e := range got {
			if strings.Contains(e.URL, want) || strings.Contains(e.Request, want) {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("EVM sources missing %q", want)
		}
	}
}

func TestSourceLogDeferredToSectionBottom(t *testing.T) {
	var b strings.Builder
	w := newWriter(&b, Options{ShowSources: true})
	w.Section("1. TEST")
	w.Subsection("Metrics")
	w.SourceLog(sampleExchanges()[:1])
	w.Row("status", "running")
	w.flush()
	out := b.String()

	sourcesIdx := strings.Index(out, `class="dash-sources"`)
	rowIdx := strings.Index(out, `class="kpi-tile"`)
	if sourcesIdx < 0 || rowIdx < 0 {
		t.Fatalf("expected deferred sources footer and KPI row in:\n%s", out)
	}
	if sourcesIdx < rowIdx {
		t.Fatal("data sources log should render after section content")
	}
	if !strings.Contains(out, `dash-sources__tag">req`) || !strings.Contains(out, `dash-sources__tag">res`) {
		t.Fatal("data sources should include raw req/res blocks")
	}
}
