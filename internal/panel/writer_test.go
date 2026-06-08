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
