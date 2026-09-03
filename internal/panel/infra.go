package panel

import (
	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func writeInfraSummary(w Writer, d model.Report, mode SummaryMode) {
	s := loadInfraState(d)
	summaryWrapStart(w, mode, "infra")
	writeInfraSummaryBody(w, d, s)
	summaryWrapEnd(w, mode)
}

func writeInfraSummaryBody(w Writer, d model.Report, s infraState) {
	w.WriteHTML(`<div class="infra-summary">`)
	if len(s.alerts) > 0 {
		w.WriteHTML(`<div class="infra-summary__hero">`)
		writeSummaryBadges(w, "infra-summary__alerts", s.alerts...)
		w.WriteHTML(`</div>`)
	}

	w.WriteHTML(`<div class="infra-summary__gauges">`)
	writeMiniGauge(w, "load", infraLoadPct(d))
	writeMiniGauge(w, "RAM", d.MemPct)
	writeMiniGauge(w, s.chainDiskLabel, s.chainDiskPct)
	w.WriteHTML(`</div>`)
	w.WriteHTML(`</div>`)
}

func writeInfra(w Writer, d model.Report) {
	s := loadInfraState(d)

	w.Section("1. INFRASTRUCTURE")
	writeEmbeddedSectionIntro(w, "Host CPU, memory, and disk pressure.")
	writeInfraSummary(w, d, SummaryEmbedded)

	w.Subsection("Host resources")
	w.WriteHTML(infraHostMetersHTML(d, s))

	writeSectionSources(w, ViewInfra, d)
	w.BlankLine()
}
