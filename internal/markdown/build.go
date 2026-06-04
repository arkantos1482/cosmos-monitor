package markdown

import (
	"strings"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

// Build renders the full operations dashboard as portable Markdown.
func Build(d model.Report) string {
	var b strings.Builder
	m := &mdWriter{w: &b}
	writeInfra(m, d)
	writeNode(m, d)
	writeValidators(m, d)
	writeLocalValidator(m, d)
	writeEconomics(m, d)
	writeGovernance(m, d)
	writeEVMSection(m, d)
	return b.String()
}
