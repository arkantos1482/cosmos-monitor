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
	w io.Writer
}

func newWriter(w io.Writer) *docWriter {
	return &docWriter{w: w}
}

func (d *docWriter) WriteString(s string) {
	fmt.Fprint(d.w, inlineHTML(s))
}

func (d *docWriter) Section(title string) {
	fmt.Fprintf(d.w, "<h1>%s</h1>\n", html.EscapeString(title))
}

func (d *docWriter) Subsection(title string) {
	fmt.Fprintf(d.w, "<h2>%s</h2>\n", html.EscapeString(title))
}

func (d *docWriter) Row(label, value string) {
	fmt.Fprintf(d.w, "<p><strong>%s</strong> %s</p>\n",
		html.EscapeString(label), inlineHTML(value))
}

func (d *docWriter) Hint(text string) {
	d.Em(text)
}

func (d *docWriter) Em(text string) {
	fmt.Fprintf(d.w, "<p><em>%s</em></p>\n", inlineHTML(text))
}

func (d *docWriter) StrongLine(text string) {
	fmt.Fprintf(d.w, "<p><strong>%s</strong></p>\n", inlineHTML(text))
}

func (d *docWriter) BlankLine() {}

func (d *docWriter) ValidatorHeader(title string) {
	fmt.Fprintf(d.w, "<p><strong>%s</strong></p>\n", inlineHTML(title))
}

func (d *docWriter) ListItem(text string) {
	fmt.Fprintf(d.w, "<ul><li>%s</li></ul>\n", inlineHTML(text))
}

func (d *docWriter) Pre(content string) {
	fmt.Fprintf(d.w, "<pre><code>%s</code></pre>\n", html.EscapeString(content))
}

func (d *docWriter) PreBash(content string) {
	fmt.Fprintf(d.w, "<pre><code class=\"language-bash\">%s</code></pre>\n", html.EscapeString(content))
}

func (d *docWriter) Mermaid(src string) {
	src = strings.TrimSpace(src)
	fmt.Fprintf(d.w, "<div class=\"mermaid\">%s</div>\n", html.EscapeString(src))
}

func (d *docWriter) MathLatex(parts ...string) {
	for _, part := range parts {
		for _, block := range splitLatexDisplayBlocks(part) {
			b64 := base64.StdEncoding.EncodeToString([]byte(block))
			fmt.Fprintf(d.w, `<div class="math-display" data-tex-b64="%s"></div>`+"\n", b64)
		}
	}
}

func (d *docWriter) Table(headers []string, rows [][]string) {
	fmt.Fprint(d.w, "<table><thead><tr>")
	for _, h := range headers {
		fmt.Fprintf(d.w, "<th>%s</th>", html.EscapeString(h))
	}
	fmt.Fprint(d.w, "</tr></thead><tbody>")
	for _, row := range rows {
		fmt.Fprint(d.w, "<tr>")
		for _, cell := range row {
			fmt.Fprintf(d.w, "<td>%s</td>", inlineHTML(cell))
		}
		fmt.Fprint(d.w, "</tr>")
	}
	fmt.Fprint(d.w, "</tbody></table>\n")
}

var (
	inlineCodeRE = regexp.MustCompile("`([^`]+)`")
	inlineBoldRE = regexp.MustCompile(`\*\*([^*]+)\*\*`)
	// Parenthetical notes only — avoids corrupting API names like node_info across <code> spans.
	inlineEmRE = regexp.MustCompile(`_\(([^)]+)\)_`)
)

func inlineHTML(s string) string {
	s = html.EscapeString(s)
	s = inlineCodeRE.ReplaceAllString(s, `<code>$1</code>`)
	s = inlineBoldRE.ReplaceAllString(s, `<strong>$1</strong>`)
	s = inlineEmRE.ReplaceAllString(s, `<em>$1</em>`)
	return s
}
