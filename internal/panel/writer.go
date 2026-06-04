package panel

import (
	"encoding/base64"
	"fmt"
	"html"
	"io"
	"regexp"
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
	Mermaid(src string)
	MathLatex(parts ...string)
	WriteString(s string)
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
	fmt.Fprint(d.w, "</dl>\n")
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
	fmt.Fprint(d.w, `<dl class="stat-grid">`+"\n")
	d.inStatGrid = true
}

func (d *docWriter) WriteString(s string) {
	d.closeList()
	d.closeStatGrid()
	fmt.Fprint(d.w, inlineHTML(s))
}

func (d *docWriter) Section(title string) {
	d.closeSection()
	fmt.Fprintf(d.w, `<section class="dash-section"><h2 class="dash-heading">%s</h2>`+"\n",
		html.EscapeString(title))
	d.inSection = true
}

func (d *docWriter) Subsection(title string) {
	d.closeSubsection()
	fmt.Fprintf(d.w, `<div class="dash-block"><h3 class="dash-subheading">%s</h3>`+"\n",
		html.EscapeString(title))
	d.inSubsection = true
}

func (d *docWriter) Row(label, value string) {
	d.openStatGrid()
	fmt.Fprintf(d.w, `<div class="stat"><dt>%s</dt><dd>%s</dd></div>`+"\n",
		html.EscapeString(label), formatValue(value))
}

func (d *docWriter) Hint(text string) {
	d.closeList()
	d.closeStatGrid()
	fmt.Fprintf(d.w, `<p class="hint">%s</p>`+"\n", inlineHTML(text))
}

func (d *docWriter) Em(text string) {
	d.closeList()
	d.closeStatGrid()
	fmt.Fprintf(d.w, `<p class="note">%s</p>`+"\n", inlineHTML(text))
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
	fmt.Fprintf(d.w, `<pre class="code-block"><code>%s</code></pre>`+"\n", html.EscapeString(content))
}

func (d *docWriter) PreBash(content string) {
	d.closeList()
	d.closeStatGrid()
	fmt.Fprintf(d.w, `<pre class="code-block"><code class="language-bash">%s</code></pre>`+"\n", html.EscapeString(content))
}

func (d *docWriter) Mermaid(src string) {
	d.closeList()
	d.closeStatGrid()
	src = strings.TrimSpace(src)
	fmt.Fprintf(d.w, `<div class="diagram-panel mermaid">%s</div>`+"\n", html.EscapeString(src))
}

func (d *docWriter) MathLatex(parts ...string) {
	d.closeList()
	d.closeStatGrid()
	for _, part := range parts {
		blocks := splitLatexDisplayBlocks(part)
		if len(blocks) == 0 {
			continue
		}
		fmt.Fprint(d.w, `<div class="math-panel">`+"\n")
		for _, block := range blocks {
			b64 := base64.StdEncoding.EncodeToString([]byte(block))
			fmt.Fprintf(d.w, `<div class="math-line" data-tex-b64="%s"></div>`+"\n", b64)
		}
		fmt.Fprint(d.w, `</div>`+"\n")
	}
}

func (d *docWriter) Table(headers []string, rows [][]string) {
	d.closeList()
	d.closeStatGrid()
	fmt.Fprint(d.w, `<div class="table-scroll"><table class="data-table"><thead><tr>`)
	for _, h := range headers {
		fmt.Fprintf(d.w, "<th>%s</th>", html.EscapeString(h))
	}
	fmt.Fprint(d.w, "</tr></thead><tbody>")
	for _, row := range rows {
		fmt.Fprint(d.w, "<tr>")
		for _, cell := range row {
			fmt.Fprintf(d.w, "<td>%s</td>", formatValue(cell))
		}
		fmt.Fprint(d.w, "</tr>")
	}
	fmt.Fprint(d.w, "</tbody></table></div>\n")
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
