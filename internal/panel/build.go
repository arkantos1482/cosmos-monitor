package panel

import (
	"strings"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

// BuildHTML renders the full operations dashboard as an HTML fragment.
func BuildHTML(d model.Report) string {
	var b strings.Builder
	writeAll(newWriter(&b, FormatHTML), d)
	return b.String()
}

// BuildText renders the full dashboard as plain text (terminal / dump).
func BuildText(d model.Report) string {
	var b strings.Builder
	writeAll(newWriter(&b, FormatText), d)
	return b.String()
}

// Build is an alias for BuildHTML (HTML panel output).
func Build(d model.Report) string {
	return BuildHTML(d)
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
