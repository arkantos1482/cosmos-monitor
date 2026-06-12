package panel

import (
	"fmt"
	"html"
	"math"

	"github.com/arkantos1482/cosmos-monitor/internal/feemarket"
	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

// SummaryMode controls whether a section summary is embedded or wrapped as an overview card link.
type SummaryMode int

const (
	SummaryEmbedded SummaryMode = iota
	SummaryOverviewClickable
)

// writeSectionLead is a one-line description shown above the embedded summary card.
func writeSectionLead(w Writer, text string) {
	if text == "" {
		return
	}
	w.Em(text)
}

// writeSectionSummaryTitle labels the summary card within a section page.
func writeSectionSummaryTitle(w Writer) {
	w.WriteHTML(`<h3 class="dash-subheading dash-section__summary-title">Summary</h3>`)
}

// writeEmbeddedSectionIntro renders lead text and the Summary heading before the card.
func writeEmbeddedSectionIntro(w Writer, lead string) {
	writeSectionLead(w, lead)
	writeSectionSummaryTitle(w)
}

func summaryWrapStart(w Writer, mode SummaryMode, slug string) {
	switch mode {
	case SummaryOverviewClickable:
		w.WriteHTML(fmt.Sprintf(`<a class="dash-overview__card dash-overview__card--%s" href="/s/%s">`,
			html.EscapeString(slug), html.EscapeString(slug)))
		if title := NavLabelForSlug(slug); title != "" {
			w.WriteHTML(fmt.Sprintf(`<p class="dash-overview__card-title">%s</p>`, html.EscapeString(title)))
		}
	case SummaryEmbedded:
		w.WriteHTML(`<div class="dash-section__summary">`)
	}
}

func summaryWrapEnd(w Writer, mode SummaryMode) {
	switch mode {
	case SummaryOverviewClickable:
		w.WriteHTML(`</a>`)
	case SummaryEmbedded:
		w.WriteHTML(`</div>`)
	}
}

type summaryBadge struct {
	text string
	kind string // ok | warn | bad
}

func writeSummaryBadges(w Writer, class string, badges ...summaryBadge) {
	if len(badges) == 0 {
		return
	}
	w.WriteHTML(`<div class="` + html.EscapeString(class) + `">`)
	for _, b := range badges {
		if b.text == "" {
			continue
		}
		cls := "badge"
		if b.kind != "" {
			cls += " badge--" + b.kind
		}
		w.WriteHTML(fmt.Sprintf(`<span class="%s">%s</span>`, cls, html.EscapeString(b.text)))
	}
	w.WriteHTML(`</div>`)
}

func writeMiniGauge(w Writer, label string, pct int) {
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	w.WriteHTML(fmt.Sprintf(
		`<div class="mini-gauge"><div class="mini-gauge__label"><span>%s</span><span>%d%%</span></div>`+
			`<div class="mini-gauge__track"><div class="mini-gauge__fill" style="width:%d%%"></div></div></div>`,
		html.EscapeString(label), pct, pct,
	))
}

func badgeKind(v string) string {
	switch badgeClass(v) {
	case "badge--ok":
		return "ok"
	case "badge--warn":
		return "warn"
	case "badge--bad":
		return "bad"
	default:
		return ""
	}
}

func feemarketBadgeKind(b feemarket.Badge) string {
	switch b.Class {
	case "rising":
		return "bad"
	case "falling", "floor":
		return "ok"
	case "disabled":
		return "warn"
	default:
		return ""
	}
}

func localBadges(d model.Report) []summaryBadge {
	if !d.Local.IsValidator {
		return nil
	}
	var b []summaryBadge
	if d.Local.Jailed {
		b = append(b, summaryBadge{"jailed", "bad"})
	}
	if d.Local.Tombstoned {
		b = append(b, summaryBadge{"tombstoned", "bad"})
	}
	if d.Local.MissedHigh {
		b = append(b, summaryBadge{"missed blocks high", "warn"})
	}
	return b
}

func pmtPoolBadge(d model.Report) summaryBadge {
	if !d.PMTEnabled {
		return summaryBadge{"PMT disabled", ""}
	}
	if d.PMTPoolEmpty {
		return summaryBadge{"pool empty", "warn"}
	}
	return summaryBadge{"PMT enabled", "ok"}
}

func loadGaugePct(load1 float64) int {
	pct := int(math.Min(load1*100/4, 100))
	if pct < 0 {
		return 0
	}
	return pct
}
