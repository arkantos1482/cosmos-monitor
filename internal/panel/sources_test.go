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
	all := sampleExchanges()
	got := exchangesForView(ViewDistribution, all)
	if len(got) != 1 || !strings.Contains(got[0].URL, "distribution") {
		t.Fatalf("distribution view should match distribution endpoint, got %d", len(got))
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
	if len(got) != 1 {
		t.Fatalf("EVM view should match jsonrpc only, got %d", len(got))
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
