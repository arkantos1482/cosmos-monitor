package panel

import (
	"fmt"
	"html"
	"strings"

	"github.com/arkantos1482/cosmos-monitor/internal/feemarket"
)

type feeLevel struct {
	ID       string
	Title    string
	Concept  string
	Badge    feemarket.Badge
	Banner   string
	Rows     [][]string
	Footnote string
	Extra    string
}

func writeFeeLevel(w Writer, lv feeLevel) {
	var b strings.Builder
	fmt.Fprintf(&b, `<details class="fee-level" id="%s" open>`, html.EscapeString(lv.ID))
	fmt.Fprintf(&b, `<summary class="fee-level__summary">%s</summary>`, html.EscapeString(lv.Title))
	b.WriteString(`<div class="fee-level__body">`)
	if lv.Concept != "" {
		fmt.Fprintf(&b, `<p class="fee-level__concept">%s</p>`, inlineHTML(lv.Concept))
	}
	if lv.Banner != "" {
		fmt.Fprintf(&b, `<p class="fee-level__banner">%s</p>`, inlineHTML(lv.Banner))
	}
	if lv.Badge.Label != "" && lv.ID == "fee-L1" {
		cls := lv.Badge.Class
		if cls == "" {
			cls = "stable"
		}
		fmt.Fprintf(&b, `<div class="fee-level__badge-row"><span class="fee-badge fee-badge--%s">%s</span></div>`,
			html.EscapeString(cls), html.EscapeString(lv.Badge.Label))
	}
	if len(lv.Rows) > 0 {
		b.WriteString(feeLevelRowsHTML(lv.Rows))
	}
	if lv.Extra != "" {
		b.WriteString(lv.Extra)
	}
	if lv.Footnote != "" {
		fmt.Fprintf(&b, `<p class="fee-level__footnote">%s</p>`, inlineHTML(lv.Footnote))
	}
	b.WriteString(`</div></details>`)
	w.WriteHTML(b.String())
}

func feeLevelRowsHTML(rows [][]string) string {
	var b strings.Builder
	b.WriteString(`<dl class="fee-level__rows">`)
	for _, row := range rows {
		if len(row) == 0 {
			continue
		}
		label := row[0]
		val := "—"
		if len(row) > 1 {
			val = row[1]
		}
		b.WriteString(`<div class="fee-level__row">`)
		fmt.Fprintf(&b, `<dt>%s</dt>`, html.EscapeString(label))
		b.WriteString(`<dd>`)
		b.WriteString(softWrapHTML(val))
		b.WriteString(`</dd></div>`)
	}
	b.WriteString(`</dl>`)
	return b.String()
}

func provenanceCalloutHTML(text string) string {
	return `<p class="dash-callout dash-callout--hint hint">` + hintHTML(text) + `</p>`
}

func noteCalloutHTML(text string) string {
	return `<p class="dash-callout dash-callout--note note">` + inlineHTML(text) + `</p>`
}

func feeSubheadingHTML(title string) string {
	return `<h3 class="dash-subheading">` + html.EscapeString(title) + `</h3>`
}

func feeTableHTML(headers []string, rows [][]string) string {
	var b strings.Builder
	b.WriteString(`<table class="fee-table"><thead><tr>`)
	for _, h := range headers {
		fmt.Fprintf(&b, `<th>%s</th>`, html.EscapeString(h))
	}
	b.WriteString(`</tr></thead><tbody>`)
	for _, row := range rows {
		b.WriteString(`<tr>`)
		for _, cell := range row {
			fmt.Fprintf(&b, `<td>%s</td>`, softWrapHTML(cell))
		}
		b.WriteString(`</tr>`)
	}
	b.WriteString(`</tbody></table>`)
	return b.String()
}
