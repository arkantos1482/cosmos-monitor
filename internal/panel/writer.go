package panel

import (
	"encoding/base64"
	"fmt"
	"html"
	"io"
	"regexp"
	"strings"
)

// Format selects HTML panel or plain terminal text.
type Format int

const (
	FormatHTML Format = iota
	FormatText
)

// Writer renders dashboard sections in HTML or plain text.
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
	w      io.Writer
	format Format
}

func newWriter(w io.Writer, format Format) *docWriter {
	return &docWriter{w: w, format: format}
}

func (d *docWriter) WriteString(s string) {
	if d.format == FormatHTML {
		fmt.Fprint(d.w, inlineHTML(s))
	} else {
		fmt.Fprint(d.w, stripInlineMD(s))
	}
}

func (d *docWriter) Section(title string) {
	if d.format == FormatHTML {
		fmt.Fprintf(d.w, "<h1>%s</h1>\n", html.EscapeString(title))
	} else {
		fmt.Fprintf(d.w, "\n# %s\n\n", title)
	}
}

func (d *docWriter) Subsection(title string) {
	if d.format == FormatHTML {
		fmt.Fprintf(d.w, "<h2>%s</h2>\n", html.EscapeString(title))
	} else {
		fmt.Fprintf(d.w, "\n## %s\n\n", title)
	}
}

func (d *docWriter) Row(label, value string) {
	if d.format == FormatHTML {
		fmt.Fprintf(d.w, "<p><strong>%s</strong> %s</p>\n",
			html.EscapeString(label), inlineHTML(value))
	} else {
		fmt.Fprintf(d.w, "- **%s**: %s\n", label, value)
	}
}

func (d *docWriter) Hint(text string) {
	d.Em(text)
}

func (d *docWriter) Em(text string) {
	if d.format == FormatHTML {
		fmt.Fprintf(d.w, "<p><em>%s</em></p>\n", inlineHTML(text))
	} else {
		fmt.Fprintf(d.w, "_%s_\n\n", text)
	}
}

func (d *docWriter) StrongLine(text string) {
	if d.format == FormatHTML {
		fmt.Fprintf(d.w, "<p><strong>%s</strong></p>\n", inlineHTML(text))
	} else {
		fmt.Fprintf(d.w, "**%s**\n\n", text)
	}
}

func (d *docWriter) BlankLine() {
	if d.format == FormatText {
		fmt.Fprintln(d.w)
	}
}

func (d *docWriter) ValidatorHeader(title string) {
	if d.format == FormatHTML {
		fmt.Fprintf(d.w, "<p><strong>%s</strong></p>\n", inlineHTML(title))
	} else {
		fmt.Fprintf(d.w, "\n%s\n\n", title)
	}
}

func (d *docWriter) ListItem(text string) {
	if d.format == FormatHTML {
		fmt.Fprintf(d.w, "<ul><li>%s</li></ul>\n", inlineHTML(text))
	} else {
		fmt.Fprintf(d.w, "- %s\n", text)
	}
}

func (d *docWriter) Pre(content string) {
	if d.format == FormatHTML {
		fmt.Fprintf(d.w, "<pre><code>%s</code></pre>\n", html.EscapeString(content))
	} else {
		fmt.Fprintf(d.w, "```text\n%s\n```\n\n", content)
	}
}

func (d *docWriter) PreBash(content string) {
	if d.format == FormatHTML {
		fmt.Fprintf(d.w, "<pre><code class=\"language-bash\">%s</code></pre>\n", html.EscapeString(content))
	} else {
		fmt.Fprintf(d.w, "```bash\n%s\n```\n\n", content)
	}
}

func (d *docWriter) Mermaid(src string) {
	src = strings.TrimSpace(src)
	if d.format == FormatHTML {
		fmt.Fprintf(d.w, "<div class=\"mermaid\">%s</div>\n", html.EscapeString(src))
	} else {
		fmt.Fprintf(d.w, "```mermaid\n%s\n```\n\n", src)
	}
}

func (d *docWriter) MathLatex(parts ...string) {
	for _, part := range parts {
		for _, block := range splitLatexDisplayBlocks(part) {
			if d.format == FormatHTML {
				b64 := base64.StdEncoding.EncodeToString([]byte(block))
				fmt.Fprintf(d.w, `<div class="math-display" data-tex-b64="%s"></div>`+"\n", b64)
			} else {
				fmt.Fprintf(d.w, "\n$$\n%s\n$$\n\n", block)
			}
		}
	}
}

func (d *docWriter) Table(headers []string, rows [][]string) {
	if d.format == FormatHTML {
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
		return
	}
	sep := make([]string, len(headers))
	for i := range sep {
		sep[i] = "---"
	}
	fmt.Fprintf(d.w, "| %s |\n", strings.Join(headers, " | "))
	fmt.Fprintf(d.w, "|%s|\n", strings.Join(sep, "|"))
	for _, row := range rows {
		fmt.Fprintf(d.w, "| %s |\n", strings.Join(row, " | "))
	}
	fmt.Fprintln(d.w)
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

func stripInlineMD(s string) string {
	s = inlineCodeRE.ReplaceAllString(s, "$1")
	s = inlineBoldRE.ReplaceAllString(s, "$1")
	s = inlineEmRE.ReplaceAllString(s, "($1)")
	return s
}
