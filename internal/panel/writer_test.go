package panel

import (
	"strings"
	"testing"
)

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
