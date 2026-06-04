package panel

import (
	"strings"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

// Build renders the full operations dashboard as an HTML fragment.
func Build(d model.Report) string {
	var b strings.Builder
	w := newWriter(&b)
	writeAll(w, d)
	w.flush()
	return b.String()
}

func writeAll(w Writer, d model.Report) {
	writeInfra(w, d)
	writeNode(w, d)
	writeValidators(w, d)
	writeLocalValidator(w, d)
	writeEconomics(w, d)
	writeGovernance(w, d)
	writeEVMSection(w, d)
}
