package panel

import (
	"fmt"
	"html"
	"math"
	"strings"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

type infraState struct {
	alerts         []summaryBadge
	chainDiskLabel string
	chainDiskPct   int
	loadDetail     string
}

func loadInfraState(d model.Report) infraState {
	s := infraState{
		chainDiskLabel: "root disk",
		chainDiskPct:   d.DiskPct,
	}
	if d.HasChainDataDisk {
		s.chainDiskLabel = "chain data"
		s.chainDiskPct = d.DataDiskPct
	}
	if tone := infraMeterTone(d.MemPct); tone != "" {
		s.alerts = append(s.alerts, summaryBadge{fmt.Sprintf("RAM %d%%", d.MemPct), tone})
	}
	if tone := infraMeterTone(s.chainDiskPct); tone != "" {
		s.alerts = append(s.alerts, summaryBadge{fmt.Sprintf("%s %d%%", s.chainDiskLabel, s.chainDiskPct), tone})
	}
	s.loadDetail = infraLoadDetail(d)
	return s
}

func infraLoadDetail(d model.Report) string {
	v := fmt.Sprintf("%.2f · %.2f · %.2f", d.Load1, d.Load5, d.Load15)
	if d.NumCPU > 0 {
		v += fmt.Sprintf("  (%d CPUs · %.2f per core @ 1m)", d.NumCPU, d.Load1/float64(d.NumCPU))
	}
	return v
}

func infraMeterTone(pct int) string {
	switch {
	case pct >= 90:
		return "bad"
	case pct >= 75:
		return "warn"
	default:
		return ""
	}
}

func infraMeterHTML(label, detail string, pct int) string {
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	fillCls := "infra-meter__fill"
	if tone := infraMeterTone(pct); tone != "" {
		fillCls += " infra-meter__fill--" + tone
	}
	var b strings.Builder
	b.WriteString(`<div class="infra-meter">`)
	fmt.Fprintf(&b, `<div class="infra-meter__head"><span class="infra-meter__label">%s</span>`,
		html.EscapeString(label))
	fmt.Fprintf(&b, `<span class="infra-meter__pct">%d%%</span></div>`, pct)
	fmt.Fprintf(&b, `<div class="infra-meter__track"><div class="%s" style="width:%d%%"></div></div>`, fillCls, pct)
	if detail != "" {
		fmt.Fprintf(&b, `<p class="infra-meter__detail">%s</p>`, html.EscapeString(detail))
	}
	b.WriteString(`</div>`)
	return b.String()
}

func infraHostMetersHTML(d model.Report, s infraState) string {
	var b strings.Builder
	b.WriteString(`<div class="infra-meters">`)
	b.WriteString(infraMeterHTML("memory", fmt.Sprintf("%s used · %s free of %s",
		d.MemUsed, orDash(d.MemAvail), d.MemTotal), d.MemPct))
	b.WriteString(infraMeterHTML(s.chainDiskLabel, diskDetailForLabel(d, s.chainDiskLabel), s.chainDiskPct))
	b.WriteString(infraMeterHTML("load (1m)", s.loadDetail, infraLoadPct(d)))
	if d.SwapTotal != "" {
		b.WriteString(infraStatHTML("swap", fmt.Sprintf("%s / %s", orDash(d.SwapUsed), d.SwapTotal)))
	}
	b.WriteString(`</div>`)
	return b.String()
}

func diskDetailForLabel(d model.Report, label string) string {
	switch label {
	case "chain data":
		return fmt.Sprintf("%s used of %s", d.DataDiskUsed, d.DataDiskTotal)
	default:
		return fmt.Sprintf("%s used · %s free of %s", d.DiskUsed, orDash(d.DiskAvail), d.DiskTotal)
	}
}

func infraStatHTML(label, detail string) string {
	return fmt.Sprintf(
		`<div class="infra-stat"><span class="infra-stat__label">%s</span>`+
			`<span class="infra-stat__value">%s</span></div>`,
		html.EscapeString(label), html.EscapeString(detail))
}

func infraLoadPerCore(d model.Report) float64 {
	if d.NumCPU <= 0 {
		return 0
	}
	return d.Load1 / float64(d.NumCPU)
}

func infraLoadPct(d model.Report) int {
	if d.NumCPU <= 0 {
		return loadGaugePct(d.Load1)
	}
	pct := int(math.Min(infraLoadPerCore(d)*100, 100))
	if pct < 0 {
		return 0
	}
	return pct
}
