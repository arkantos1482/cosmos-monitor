package markdown

import (
	"fmt"
	"io"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

type mdWriter struct {
	w io.Writer
}

func (m *mdWriter) section(name string)     { fmt.Fprintf(m.w, "\n# %s\n\n", name) }
func (m *mdWriter) subsection(name string)  { fmt.Fprintf(m.w, "\n## %s\n\n", name) }
func (m *mdWriter) row(label, value string) { fmt.Fprintf(m.w, "- **%s**: %s\n", label, value) }
func (m *mdWriter) hint(text string)        { fmt.Fprintf(m.w, "_%s_\n\n", text) }

func valLocalMark(v model.Validator) string {
	if v.IsLocal {
		return "**this node**"
	}
	return ""
}

func mdPMTStatus(d model.Report) string {
	if !d.PMTEnabled {
		return "disabled"
	}
	if d.PMTPoolEmpty {
		return "enabled — pool empty  (no PMT rewards distributing)"
	}
	suffix := ""
	if d.PMTRunway != "" {
		suffix = "  (" + d.PMTRunway + ")"
	}
	return "distributing  " + d.PMTRate + "   pool " + d.PMTBalance + suffix
}
