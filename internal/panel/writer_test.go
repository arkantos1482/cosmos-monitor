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
	w.Hint("`moniker`, `node ID`, `version`, `chain ID`, `p2p listen`, `rpc listen` → CometBFT RPC `GET /status` (`node_info`, `sync_info`, `validator_info` not used here).")
	w.flush()
	out := b.String()

	for _, want := range []string{
		`class="dash-callout dash-callout--hint hint"`,
		`class="hint-provenance"`,
		`class="hint-provenance__chip"`,
		`class="hint-provenance__arrow"`,
		`class="hint-provenance__source"`,
		`<code>moniker</code>`,
		`<code>GET /status</code>`,
		`<code>validator_info</code> not used here`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in:\n%s", want, out)
		}
	}
}

func TestHintProvenanceMultiClause(t *testing.T) {
	html := hintHTML("`load` → `/proc/loadavg`; `ram` → `/proc/meminfo` (MemTotal, MemAvailable); `disk` → `statfs` on `/`.")
	if !strings.Contains(html, `hint-provenance__sep`) {
		t.Fatalf("expected clause separator in:\n%s", html)
	}
	if strings.Count(html, `hint-provenance__clause`) != 3 {
		t.Fatalf("expected 3 clauses, got:\n%s", html)
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

func TestHintFallbackChainedArrows(t *testing.T) {
	text := "Follows BeginBlock: sources → `fee_collector` → community tax + validator pool → operators and delegators."
	html := hintHTML(text)
	if strings.Contains(html, "hint-provenance") {
		t.Fatal("chained-arrow hint should fall back to inline markup")
	}
}
