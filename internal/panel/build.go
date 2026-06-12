package panel

import (
	"strings"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

// Build renders all seven sections as one HTML fragment (used by --dump).
func Build(d model.Report) string {
	return BuildWithOptions(d, Options{})
}

// BuildWithOptions renders all sections with optional display flags.
func BuildWithOptions(d model.Report, opts Options) string {
	var b strings.Builder
	w := newWriter(&b, opts)
	writeAll(w, d)
	w.flush()
	return b.String()
}

// BuildView renders the home overview or a single section fragment.
func BuildView(v View, d model.Report) string {
	return BuildViewWithOptions(v, d, Options{})
}

// BuildViewWithOptions renders one view with optional display flags.
func BuildViewWithOptions(v View, d model.Report, opts Options) string {
	var b strings.Builder
	w := newWriter(&b, opts)
	writeView(w, v, d)
	w.flush()
	return b.String()
}

func writeAll(w Writer, d model.Report) {
	writeStaking(w, d)
	writeRewards(w, d)
	writeEconomics(w, d)
	writeFeemarket(w, d)
	writeGovernance(w, d)
	writeInfra(w, d)
	writeNode(w, d)
	writeEVMSection(w, d)
}
