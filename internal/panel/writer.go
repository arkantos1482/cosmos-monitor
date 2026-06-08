package panel

import (
	"fmt"
	"html"
	"io"
	"regexp"
	"strconv"
	"strings"
)

// Writer renders dashboard sections as an HTML fragment.
type Writer interface {
	Section(title string)
	Subsection(title string)
	Row(label, value string)
	Hint(text string)
	Em(text string)
	StrongLine(text string)
	BlankLine()
	Table(headers []string, rows [][]string)
	ValidatorHeader(title string)
	ListItem(text string)
	Pre(content string)
	PreBash(content string)
	WriteString(s string)
	WriteHTML(s string) // trusted panel markup (not escaped)
	Details(id, summary string, fn func(Writer))
}

type docWriter struct {
	w            io.Writer
	inSection    bool
	inSubsection bool
	inStatGrid   bool
	inList       bool
}

func newWriter(w io.Writer) *docWriter {
	return &docWriter{w: w}
}

func (d *docWriter) flush() {
	d.closeList()
	d.closeStatGrid()
	d.closeSubsection()
	d.closeSection()
}

func (d *docWriter) closeList() {
	if !d.inList {
		return
	}
	fmt.Fprint(d.w, "</ul>\n")
	d.inList = false
}

func (d *docWriter) closeStatGrid() {
	if !d.inStatGrid {
		return
	}
	fmt.Fprint(d.w, "</div>\n")
	d.inStatGrid = false
}

func (d *docWriter) closeSubsection() {
	d.closeList()
	d.closeStatGrid()
	if !d.inSubsection {
		return
	}
	fmt.Fprint(d.w, "</div>\n")
	d.inSubsection = false
}

func (d *docWriter) closeSection() {
	d.closeSubsection()
	if !d.inSection {
		return
	}
	fmt.Fprint(d.w, "</section>\n")
	d.inSection = false
}

func (d *docWriter) openStatGrid() {
	if d.inStatGrid {
		return
	}
	d.closeList()
	fmt.Fprint(d.w, `<div class="kpi-grid">`+"\n")
	d.inStatGrid = true
}

func (d *docWriter) WriteString(s string) {
	d.closeList()
	d.closeStatGrid()
	fmt.Fprint(d.w, inlineHTML(s))
}

func (d *docWriter) WriteHTML(s string) {
	d.closeList()
	d.closeStatGrid()
	fmt.Fprint(d.w, s)
}

func sectionSlug(title string) string {
	upper := strings.ToUpper(title)
	switch {
	case strings.Contains(upper, "INFRASTRUCTURE"):
		return "infra"
	case strings.Contains(upper, "NODE") && !strings.Contains(upper, "VALIDATOR"):
		return "node"
	case strings.Contains(upper, "VALIDATOR SET"):
		return "validators"
	case strings.Contains(upper, "THIS VALIDATOR"):
		return "local"
	case strings.Contains(upper, "ECONOMICS"):
		return "economics"
	case strings.Contains(upper, "GOVERNANCE"):
		return "governance"
	case strings.Contains(upper, "EVM"):
		return "evm"
	default:
		return ""
	}
}

func (d *docWriter) Section(title string) {
	d.closeSection()
	slug := sectionSlug(title)
	cls := "dash-section"
	if slug != "" {
		cls += " dash-section--" + slug
	}
	fmt.Fprintf(d.w, `<section class="%s"><h2 class="dash-heading">%s</h2>`+"\n",
		cls, html.EscapeString(title))
	d.inSection = true
}

func (d *docWriter) Subsection(title string) {
	d.closeSubsection()
	fmt.Fprintf(d.w, `<div class="dash-block"><div class="dash-block__header"><h3 class="dash-subheading">%s</h3></div>`+"\n",
		html.EscapeString(title))
	d.inSubsection = true
}

var (
	pctInValueRE  = regexp.MustCompile(`\((\d+(?:\.\d+)?)%\)`)
	longHexValueRE = regexp.MustCompile(`^[0-9a-fA-F]{32,}$`)
)

func (d *docWriter) Row(label, value string) {
	d.openStatGrid()
	tileClass, valHTML := kpiValueHTML(value)
	barHTML := kpiBarHTML(value)
	fmt.Fprintf(d.w, `<div class="kpi-tile%s"><div class="kpi-tile__label">%s</div><div class="kpi-tile__value">%s</div>%s</div>`+"\n",
		tileClass, html.EscapeString(label), valHTML, barHTML)
}

// kpiValueHTML structures long identifiers and _(caption)_ suffixes into
// detail/hash tile layouts; short plain values keep the inline format.
func kpiValueHTML(value string) (tileClass string, htmlOut string) {
	primary, caption := splitKPICaption(value)
	if caption == "" && !isLongKPIPrimary(primary) {
		return "", formatValue(value)
	}
	var classes []string
	if caption != "" {
		classes = append(classes, "kpi-tile--detail")
	}
	if isLongKPIPrimary(primary) {
		classes = append(classes, "kpi-tile--hash")
	}
	body := formatValue(primary)
	if isLongKPIPrimary(primary) {
		body = `<span class="kpi-tile__primary">` + body + `</span>`
	}
	if caption != "" {
		body += `<div class="kpi-tile__caption">` + inlineHTML(caption) + `</div>`
	}
	if len(classes) == 0 {
		return "", body
	}
	return " " + strings.Join(classes, " "), body
}

func splitKPICaption(s string) (primary, caption string) {
	loc := inlineEmRE.FindStringSubmatchIndex(s)
	if loc == nil {
		return s, ""
	}
	primary = strings.TrimSpace(s[:loc[0]])
	caption = s[loc[2]:loc[3]]
	return primary, caption
}

func isLongKPIPrimary(s string) bool {
	plain := strings.TrimSpace(stripInlineMarkup(s))
	if plain == "" {
		return false
	}
	if longHexValueRE.MatchString(plain) {
		return true
	}
	if strings.HasPrefix(plain, "tcp://") || strings.HasPrefix(plain, "http://") {
		return true
	}
	if strings.HasPrefix(plain, "cosmos") || strings.HasPrefix(plain, "0x") {
		return len(plain) >= 20
	}
	if strings.Contains(plain, "@") {
		return true
	}
	if strings.Count(plain, ":") >= 2 && len(plain) > 28 {
		return true
	}
	// opaque ids (no spaces); skip prose sentences
	if !strings.Contains(plain, " ") && len(plain) >= 40 {
		return true
	}
	return false
}

func stripInlineMarkup(s string) string {
	s = inlineCodeRE.ReplaceAllString(s, "$1")
	s = inlineBoldRE.ReplaceAllString(s, "$1")
	s = inlineEmRE.ReplaceAllString(s, "$1")
	return s
}

func kpiBarHTML(value string) string {
	m := pctInValueRE.FindStringSubmatch(value)
	if len(m) < 2 {
		return ""
	}
	pct, err := strconv.ParseFloat(m[1], 64)
	if err != nil || pct < 0 {
		return ""
	}
	if pct > 100 {
		pct = 100
	}
	return fmt.Sprintf(`<div class="kpi-bar"><div class="kpi-bar__fill" style="width:%.1f%%"></div></div>`, pct)
}

func (d *docWriter) Hint(text string) {
	d.closeList()
	d.closeStatGrid()
	fmt.Fprintf(d.w, `<p class="dash-callout dash-callout--hint hint">%s</p>`+"\n", hintHTML(text))
}

func (d *docWriter) Em(text string) {
	d.closeList()
	d.closeStatGrid()
	fmt.Fprintf(d.w, `<p class="dash-callout dash-callout--note note">%s</p>`+"\n", inlineHTML(text))
}

func (d *docWriter) StrongLine(text string) {
	d.closeList()
	d.closeStatGrid()
	fmt.Fprintf(d.w, `<p class="callout">%s</p>`+"\n", inlineHTML(text))
}

func (d *docWriter) BlankLine() {}

func (d *docWriter) ValidatorHeader(title string) {
	d.closeList()
	d.closeStatGrid()
	fmt.Fprintf(d.w, `<p class="validator-label">%s</p>`+"\n", inlineHTML(title))
}

func (d *docWriter) ListItem(text string) {
	d.closeStatGrid()
	if !d.inList {
		fmt.Fprint(d.w, `<ul class="dash-list">`+"\n")
		d.inList = true
	}
	fmt.Fprintf(d.w, "<li>%s</li>\n", inlineHTML(text))
}

func (d *docWriter) Pre(content string) {
	d.closeList()
	d.closeStatGrid()
	fmt.Fprintf(d.w, `<pre class="code-block terminal-panel"><code>%s</code></pre>`+"\n", html.EscapeString(content))
}

func (d *docWriter) PreBash(content string) {
	d.closeList()
	d.closeStatGrid()
	fmt.Fprintf(d.w, `<pre class="code-block terminal-panel"><code class="language-bash">%s</code></pre>`+"\n", html.EscapeString(content))
}

func (d *docWriter) Details(id, summary string, fn func(Writer)) {
	d.closeList()
	d.closeStatGrid()
	if id != "" {
		fmt.Fprintf(d.w, `<details class="dash-details" id="%s" data-details-key="%s" hx-preserve><summary class="dash-details__summary">%s</summary><div class="dash-details__body">`+"\n",
			html.EscapeString(id), html.EscapeString(id), html.EscapeString(summary))
	} else {
		fmt.Fprintf(d.w, `<details class="dash-details" hx-preserve><summary class="dash-details__summary">%s</summary><div class="dash-details__body">`+"\n",
			html.EscapeString(summary))
	}
	if fn != nil {
		fn(d)
	}
	fmt.Fprint(d.w, `</div></details>`+"\n")
}

var numericCellRE = regexp.MustCompile(`^[\d,.\s%+\-]+$`)

func (d *docWriter) Table(headers []string, rows [][]string) {
	d.closeList()
	d.closeStatGrid()
	ledger := len(headers) > 0 && headers[0] == "Step"
	tableCls := "data-table"
	if ledger {
		tableCls += " data-table--ledger"
	}
	fmt.Fprintf(d.w, `<div class="table-scroll"><table class="%s"><thead><tr>`, tableCls)
	for i, h := range headers {
		thCls := ""
		if i > 0 && isNumericHeader(h) {
			thCls = ` class="data-table__num"`
		}
		fmt.Fprintf(d.w, "<th%s>%s</th>", thCls, html.EscapeString(h))
	}
	fmt.Fprint(d.w, "</tr></thead><tbody>")
	for _, row := range rows {
		fmt.Fprint(d.w, "<tr>")
		for i, cell := range row {
			if ledger && i == 0 {
				step := html.EscapeString(strings.TrimSpace(cell))
				fmt.Fprintf(d.w, `<td class="data-table__step" data-step="%s">%s</td>`, step, step)
				continue
			}
			tdCls := ""
			if i > 0 && looksNumeric(cell) {
				tdCls = ` class="data-table__num"`
			}
			fmt.Fprintf(d.w, "<td%s>%s</td>", tdCls, formatValue(cell))
		}
		fmt.Fprint(d.w, "</tr>")
	}
	fmt.Fprint(d.w, "</tbody></table></div>\n")
}

func isNumericHeader(h string) bool {
	lower := strings.ToLower(h)
	return strings.Contains(lower, "balance") ||
		strings.Contains(lower, "block") && !strings.Contains(lower, "where") ||
		lower == "value" || lower == "check"
}

func looksNumeric(s string) bool {
	plain := strings.TrimSpace(s)
	if plain == "" || plain == "—" {
		return false
	}
	// strip HTML badge content check on raw string
	if strings.HasPrefix(plain, "0.") || strings.HasPrefix(plain, "~") {
		return true
	}
	return numericCellRE.MatchString(plain)
}

func formatValue(s string) string {
	if cls := badgeClass(s); cls != "" {
		return fmt.Sprintf(`<span class="badge %s">%s</span>`, cls, inlineHTML(s))
	}
	return inlineHTML(s)
}

func badgeClass(value string) string {
	v := strings.ToLower(strings.TrimSpace(value))
	if i := strings.IndexByte(v, ' '); i > 0 {
		v = v[:i]
	}
	switch v {
	case "running", "synced", "ok", "true", "healthy", "up":
		return "badge--ok"
	case "stopped", "down", "false", "jailed", "tombstoned":
		return "badge--bad"
	case "catching", "syncing", "degraded", "warn", "warning":
		return "badge--warn"
	}
	if strings.Contains(v, "catching") || strings.Contains(v, "degraded") {
		return "badge--warn"
	}
	return ""
}

var (
	inlineCodeRE = regexp.MustCompile("`([^`]+)`")
	inlineBoldRE = regexp.MustCompile(`\*\*([^*]+)\*\*`)
	inlineEmRE   = regexp.MustCompile(`_\(([^)]+)\)_`)
)

func inlineHTML(s string) string {
	s = html.EscapeString(s)
	s = inlineCodeRE.ReplaceAllString(s, `<code>$1</code>`)
	s = inlineBoldRE.ReplaceAllString(s, `<strong>$1</strong>`)
	s = inlineEmRE.ReplaceAllString(s, `<em>$1</em>`)
	return s
}

// hintHTML renders data-source hints. When the text matches the common
// "field list → API/source" pattern it becomes structured provenance markup;
// otherwise it falls back to inline callout text.
func hintHTML(text string) string {
	if !isProvenanceHint(text) {
		return inlineHTML(text)
	}
	clauses := splitHintClauses(text)
	var b strings.Builder
	b.WriteString(`<span class="hint-provenance">`)
	for _, clause := range clauses {
		writeHintClause(&b, strings.TrimSpace(clause))
	}
	b.WriteString(`</span>`)
	return b.String()
}

func isProvenanceHint(text string) bool {
	if !strings.Contains(text, " → ") {
		return false
	}
	for _, clause := range splitHintClauses(text) {
		idx := strings.Index(clause, " → ")
		if idx < 0 {
			return false
		}
		if strings.Contains(clause[idx+len(" → "):], " → ") {
			return false
		}
	}
	return true
}

func splitHintClauses(text string) []string {
	parts := strings.Split(text, "; ")
	var clauses []string
	var cur strings.Builder
	for i, part := range parts {
		if cur.Len() > 0 {
			cur.WriteString("; ")
		}
		cur.WriteString(part)
		if strings.Contains(part, " → ") || i == len(parts)-1 {
			clauses = append(clauses, cur.String())
			cur.Reset()
		}
	}
	if cur.Len() > 0 {
		clauses = append(clauses, cur.String())
	}
	return clauses
}

func writeHintClause(b *strings.Builder, clause string) {
	idx := strings.Index(clause, " → ")
	if idx < 0 {
		b.WriteString(inlineHTML(clause))
		return
	}
	left := strings.TrimSpace(clause[:idx])
	right := strings.TrimSpace(clause[idx+len(" → "):])
	b.WriteString(`<span class="hint-provenance__clause">`)
	b.WriteString(`<span class="hint-provenance__fields">`)
	writeHintFields(b, left)
	b.WriteString(`</span>`)
	b.WriteString(`<span class="hint-provenance__arrow" aria-hidden="true">→</span>`)
	b.WriteString(`<span class="hint-provenance__source">`)
	b.WriteString(inlineHTML(right))
	b.WriteString(`</span></span>`)
}

func writeHintFields(b *strings.Builder, left string) {
	fields := extractBacktickFields(left)
	if len(fields) > 0 {
		if prefix := hintFieldPrefix(left, fields); prefix != "" {
			fmt.Fprintf(b, `<span class="hint-provenance__prefix">%s</span>`, inlineHTML(prefix))
		}
		for _, f := range fields {
			fmt.Fprintf(b, `<span class="hint-provenance__chip"><code>%s</code></span>`, html.EscapeString(f))
		}
		return
	}
	trimmed := strings.TrimSpace(left)
	if trimmed == "" {
		return
	}
	parts := splitPlainFieldList(trimmed)
	if len(parts) > 1 {
		for _, p := range parts {
			fmt.Fprintf(b, `<span class="hint-provenance__chip"><code>%s</code></span>`, html.EscapeString(p))
		}
		return
	}
	fmt.Fprintf(b, `<span class="hint-provenance__prefix">%s</span>`, inlineHTML(trimmed))
}

func extractBacktickFields(s string) []string {
	var fields []string
	for _, m := range inlineCodeRE.FindAllStringSubmatch(s, -1) {
		fields = append(fields, m[1])
	}
	return fields
}

func hintFieldPrefix(left string, fields []string) string {
	s := left
	for _, f := range fields {
		s = strings.Replace(s, "`"+f+"`", "", 1)
	}
	s = strings.TrimSpace(s)
	s = strings.Trim(s, "/,&, ")
	s = strings.TrimSpace(strings.TrimSuffix(s, ":"))
	if s == "" {
		return ""
	}
	if !strings.HasSuffix(s, ":") {
		s += ":"
	}
	return s + " "
}

func splitPlainFieldList(s string) []string {
	if strings.Contains(s, " / ") {
		var parts []string
		for _, p := range strings.Split(s, " / ") {
			if t := strings.TrimSpace(p); t != "" {
				parts = append(parts, t)
			}
		}
		return parts
	}
	if strings.Contains(s, ", ") {
		var parts []string
		for _, p := range strings.Split(s, ", ") {
			if t := strings.TrimSpace(p); t != "" {
				parts = append(parts, t)
			}
		}
		return parts
	}
	return nil
}
