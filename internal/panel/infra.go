package panel

import (
	"fmt"
	"html"

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
	w.WriteHTML(`<div class="infra-summary__hero">`)
	w.WriteHTML(`<div class="infra-summary__identity">`)
	writeSummaryBadges(w, "infra-summary__status", s.containerBadge)
	if s.imageShort != "" {
		w.WriteHTML(fmt.Sprintf(`<code class="infra-summary__image">%s</code>`, html.EscapeString(s.imageShort)))
	}
	w.WriteHTML(`</div>`)
	if len(s.alerts) > 0 {
		writeSummaryBadges(w, "infra-summary__alerts", s.alerts...)
	}
	w.WriteHTML(`</div>`)

	w.WriteHTML(`<div class="infra-summary__gauges">`)
	writeMiniGauge(w, "load", infraLoadPct(d))
	writeMiniGauge(w, "RAM", d.MemPct)
	writeMiniGauge(w, s.chainDiskLabel, s.chainDiskPct)
	if s.chainDiskLabel == "chain data" {
		writeMiniGauge(w, "root disk", d.DiskPct)
	}
	w.WriteHTML(`</div>`)

	if foot := infraSummaryFootHTML(d); foot != "" {
		w.WriteHTML(fmt.Sprintf(`<p class="infra-summary__foot">%s</p>`, foot))
	}
	w.WriteHTML(`</div>`)
}

func writeInfra(w Writer, d model.Report) {
	s := loadInfraState(d)

	w.Section("1. INFRASTRUCTURE")
	writeEmbeddedSectionIntro(w, "Host CPU, memory, and disk pressure — plus the `evmd-node` Docker container that runs this validator.")
	writeInfraSummary(w, d, SummaryEmbedded)

	w.Subsection("Host resources")
	w.WriteHTML(infraHostMetersHTML(d, s))

	w.Subsection("Container")
	w.WriteHTML(`<div class="eco-domains">`)
	w.WriteHTML(infraContainerCardHTML(d, s))
	w.WriteHTML(`</div>`)

	writeSectionSources(w, ViewInfra, d)
	w.BlankLine()
}
