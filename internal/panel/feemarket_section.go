package panel

import (
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

func writeFeemarketSection(w Writer, d model.Report) {
	ex := buildFeemarketExplain(d)

	w.StrongLine(ex.SummaryLine)
	w.Em("Plain-language steps first, then the full x/feemarket formulas and live substitution with your chain’s parameters. All numbers are from this node.")
	w.MathLatex(ex.LatexGeneral, ex.LatexSubstituted)
	writeFeemarketDiagram(w, d)
}
