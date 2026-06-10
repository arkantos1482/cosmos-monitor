package panel

import (
	"strings"
	"testing"
)

func TestReferenceTableSoftWrap(t *testing.T) {
	var b strings.Builder
	w := newWriter(&b)
	w.Table([]string{"Symbol", "Value", "Meaning"}, [][]string{
		{
			"target",
			"MaxUint64 ÷ 2 (sentinel)",
			"gasLimit ÷ elasticity; max_gas = −1 → gasLimit = MaxUint64",
		},
	})
	out := b.String()
	if !strings.Contains(out, `<td class="data-table__val">MaxUint64 ÷ 2 (sentinel)</td>`) {
		t.Fatalf("value cell should stay intact on one line:\n%s", out)
	}
	for _, frag := range []string{";\u200b ", " →\u200b ", " ÷\u200b "} {
		if !strings.Contains(out, frag) {
			t.Fatalf("meaning cell missing soft-break %q in:\n%s", frag, out)
		}
	}
}

func TestReferenceTableAlignment(t *testing.T) {
	var b strings.Builder
	w := newWriter(&b)
	w.Table([]string{"Setting", "Value", "Meaning"}, [][]string{
		{"elasticity_multiplier", "2", "Target = gasLimit ÷ elasticity"},
		{"no_base_fee", "no", "EIP-1559 auto-adjust enabled"},
	})
	// legacy column order is normalized to reference | value | meaning
	w.Table([]string{"Symbol", "Meaning", "Live value"}, [][]string{
		{"base", "Base fee this block", "0.000000000000000007 PMT"},
	})
	w.Table([]string{"Parameter", "Description", "Current"}, [][]string{
		{"min_gas_multiplier", "mempool gas × multiplier", "0.5"},
	})
	out := b.String()
	for _, want := range []string{
		`data-table--reference`,
		`class="data-table__val"`,
		`class="data-table__desc"`,
		`class="data-table__key"`,
		`<th class="data-table__key">Symbol</th>`,
		`<th class="data-table__val">Value</th>`,
		`<th class="data-table__desc">Meaning</th>`,
		`<th class="data-table__key">Parameter</th>`,
		`<td class="data-table__val">0.000000000000000007 PMT</td>`,
		`<td class="data-table__desc">Base fee this block</td>`,
		`<td class="data-table__val">0.5</td>`,
		`<td class="data-table__desc">mempool gas ×​ multiplier</td>`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in:\n%s", want, out)
		}
	}
	if strings.Contains(out, `Live value`) {
		t.Fatal("reference tables should use canonical Value header")
	}
	if strings.Contains(out, `data-table__num`) {
		t.Fatalf("reference tables must not use numeric right-align:\n%s", out)
	}
}

func TestWriterComponents(t *testing.T) {
	var b strings.Builder
	w := newWriter(&b)
	w.Section("1. TEST")
	w.Subsection("Metrics")
	w.Row("status", "running")
	w.Row("ram", "4G / 8G  (50%)")
	w.ListItem("peer A")
	w.ListItem("peer B")
	w.Table([]string{"Step", "Where", "Value"}, [][]string{{"1", "fee_collector", "0.1 PMT"}})
	w.flush()
	out := b.String()

	for _, want := range []string{
		`<section class="dash-section">`,
		`dash-block__header`,
		`<div class="kpi-grid">`,
		`<div class="kpi-tile">`,
		`kpi-bar`,
		`<span class="badge badge--ok">running</span>`,
		`<ul class="dash-list">`,
		`<li>peer A</li>`,
		`<li>peer B</li>`,
		`</ul>`,
		`data-table--ledger`,
		`data-table__step`,
		`</section>`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in:\n%s", want, out)
		}
	}
	if strings.Count(out, "<ul") != 1 {
		t.Fatalf("expected one list, got %d <ul tags", strings.Count(out, "<ul"))
	}
}

func TestHintProvenanceMarkup(t *testing.T) {
	var b strings.Builder
	w := newWriter(&b)
	w.Hint("`moniker`, `node ID`, `version`, `chain ID`, `p2p listen`, `rpc listen` → CometBFT GET /status (node_info only; sync_info and validator_info in Consensus).")
	w.flush()
	out := b.String()

	for _, want := range []string{
		`class="dash-callout dash-callout--hint hint"`,
		`class="hint-provenance"`,
		`class="hint-provenance__chip"`,
		`class="hint-provenance__arrow"`,
		`class="hint-provenance__source"`,
		`<code>moniker</code>`,
		`CometBFT GET /status`,
		`validator_info in Consensus`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in:\n%s", want, out)
		}
	}
}

func TestHintProvenanceMultiClause(t *testing.T) {
	html := hintHTML("`load` → proc /proc/loadavg; `ram` → proc /proc/meminfo (MemTotal, MemAvailable); `disk` → fs statfs /.")
	if strings.Count(html, `hint-provenance__clause`) != 3 {
		t.Fatalf("expected 3 clauses, got:\n%s", html)
	}
	if strings.Contains(html, `hint-provenance__sep`) {
		t.Fatal("vertical hints should not use inline clause separators")
	}
}

func TestHintFallbackNoArrow(t *testing.T) {
	text := "Live REST balances and per-block rates."
	html := hintHTML(text)
	if strings.Contains(html, "hint-provenance") {
		t.Fatal("plain hint should not use provenance markup")
	}
	if html != inlineHTML(text) {
		t.Fatalf("expected inline fallback, got:\n%s", html)
	}
}

func TestKPIHashDetailTile(t *testing.T) {
	var b strings.Builder
	w := newWriter(&b)
	w.Row("node ID", "3381ddd6b06ec766400d3bdbddcfaaa2305f4984  _(CometBFT P2P peer ID)_")
	w.flush()
	out := b.String()

	for _, want := range []string{
		`class="kpi-tile kpi-tile--detail kpi-tile--hash"`,
		`class="kpi-tile__primary"`,
		`class="kpi-tile__caption"`,
		`3381ddd6b06ec766400d3bdbddcfaaa2305f4984`,
		`CometBFT P2P peer ID`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in:\n%s", want, out)
		}
	}
}

func TestKPIPlainTileUnchanged(t *testing.T) {
	var b strings.Builder
	w := newWriter(&b)
	w.Row("height", "482,160")
	w.flush()
	out := b.String()
	if strings.Contains(out, "kpi-tile--detail") || strings.Contains(out, "kpi-tile--hash") {
		t.Fatalf("short plain value should not use detail/hash classes:\n%s", out)
	}
}

func TestHintFallbackChainedArrows(t *testing.T) {
	text := "Follows BeginBlock: sources → `fee_collector` → community tax + validator pool → operators and delegators."
	html := hintHTML(text)
	if strings.Contains(html, "hint-provenance") {
		t.Fatal("chained-arrow hint should fall back to inline markup")
	}
}

func TestHintProvenanceSemicolonInsideParens(t *testing.T) {
	html := hintHTML("`moniker` → CometBFT GET /status (node_info only; sync_info in Consensus).")
	if !strings.Contains(html, `hint-provenance`) {
		t.Fatalf("semicolon inside parentheses must not split clauses:\n%s", html)
	}
}

func TestHintProvenancePMTRewards(t *testing.T) {
	html := hintHTML("`status`, `pool address` → REST GET /cosmos/evm/pmtrewards/v1/params; `per-block rate`, `pool balance` → ledger (Block reward ledger above).")
	if strings.Count(html, `hint-provenance__clause`) != 2 {
		t.Fatalf("expected 2 provenance clauses for PMT rewards hint, got:\n%s", html)
	}
	if !strings.Contains(html, `hint-provenance`) {
		t.Fatal("PMT rewards hint must use provenance markup, not inline fallback")
	}
}

func TestHintProvenanceVerticalClauses(t *testing.T) {
	var b strings.Builder
	w := newWriter(&b)
	w.Hint("`load` → proc /proc/loadavg; `ram` → proc /proc/meminfo (MemTotal, MemAvailable); `disk` → fs statfs /.")
	w.flush()
	out := b.String()
	if !strings.Contains(out, `class="dash-callout dash-callout--hint hint"`) {
		t.Fatalf("expected hint callout wrapper in:\n%s", out)
	}
}

func TestSectionHintsHTMLPerHintParsing(t *testing.T) {
	hints := []string{
		"`p2p listen`, `p2p dial`, `rpc listen` → CometBFT GET /status (node_info; dial is node_id@listen_addr).",
		"`status`, `jailed` → REST GET /cosmos/staking/v1beta1/validators.",
	}
	html := sectionHintsHTML(hints)
	if !strings.Contains(html, `hint-provenance`) {
		t.Fatalf("expected provenance markup:\n%s", html)
	}
	if strings.Count(html, `hint-provenance__clause`) != 2 {
		t.Fatalf("expected 2 clauses from independent hints, got:\n%s", html)
	}
}

func TestHintDeferredToSectionBottom(t *testing.T) {
	var b strings.Builder
	w := newWriter(&b)
	w.Section("1. TEST")
	w.Subsection("Metrics")
	w.Hint("`status` → docker GET /containers/{name}/json.")
	w.Row("status", "running")
	w.flush()
	out := b.String()

	hintIdx := strings.Index(out, `class="dash-sources"`)
	rowIdx := strings.Index(out, `class="kpi-tile`)
	if hintIdx < 0 || rowIdx < 0 {
		t.Fatalf("expected deferred sources footer and KPI row in:\n%s", out)
	}
	if hintIdx < rowIdx {
		t.Fatal("data sources hint should render after section content")
	}
	if !strings.Contains(out, `>Data sources</h3>`) {
		t.Fatal("deferred hints should use Data sources heading")
	}
}
