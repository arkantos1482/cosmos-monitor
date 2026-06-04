package markdown

import (
	"fmt"
	"io"
	"strings"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func splitLatexDisplayBlocks(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	var blocks []string
	for _, chunk := range strings.Split(s, `\]`) {
		chunk = strings.TrimSpace(chunk)
		chunk = strings.TrimPrefix(chunk, `\[`)
		chunk = strings.TrimSpace(chunk)
		if chunk != "" {
			blocks = append(blocks, chunk)
		}
	}
	return blocks
}

func writeFeeMathMarkdown(w io.Writer, latexParts ...string) {
	for _, part := range latexParts {
		for _, block := range splitLatexDisplayBlocks(part) {
			fmt.Fprintf(w, "\n$$\n%s\n$$\n\n", block)
		}
	}
}

func writeFeemarketSection(w io.Writer, d model.Report) {
	ex := buildFeemarketExplain(d)

	fmt.Fprintf(w, "**%s**\n\n", ex.SummaryLine)
	writeFeeMathMarkdown(w, ex.LatexGeneral, ex.LatexSubstituted)
	writeFeemarketDiagram(w, d)
}
