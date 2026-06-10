package panel

import (
	"strings"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

// Build renders all seven sections as one HTML fragment (used by --dump).
func Build(d model.Report) string {
	var b strings.Builder
	w := newWriter(&b)
	writeAll(w, d)
	w.flush()
	return b.String()
}

// BuildView renders the home overview or a single section fragment.
func BuildView(v View, d model.Report) string {
	var b strings.Builder
	w := newWriter(&b)
	writeView(w, v, d)
	w.flush()
	return b.String()
}

func writeAll(w Writer, d model.Report) {
	writeInfra(w, d)
	writeNode(w, d)
	writeValidators(w, d)
	writeEconomics(w, d)
	writeFeemarket(w, d)
	writeGovernance(w, d)
	writeEVMSection(w, d)
}
