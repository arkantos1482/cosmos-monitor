package panel

import (
	"fmt"
	"html"
	"math"
	"strings"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

type infraState struct {
	imageShort     string
	containerBadge summaryBadge
	alerts         []summaryBadge
	chainDiskLabel string
	chainDiskPct   int
	loadDetail     string
}

func loadInfraState(d model.Report) infraState {
	s := infraState{
		imageShort:     infraImageShort(d.NodeImage),
		chainDiskLabel: "root disk",
		chainDiskPct:   d.DiskPct,
	}
	if d.NodeRunning {
		s.containerBadge = summaryBadge{"running", "ok"}
	} else {
		s.containerBadge = summaryBadge{"stopped", "bad"}
	}
	if d.NodeOOMKilled {
		s.alerts = append(s.alerts, summaryBadge{"OOM killed", "bad"})
	}
	if d.Restarts > 0 {
		kind := ""
		if d.Restarts >= 3 {
			kind = "warn"
		}
		s.alerts = append(s.alerts, summaryBadge{fmt.Sprintf("%d restarts", d.Restarts), kind})
	}
	if d.DataPath != "" && d.DataDiskPct > 0 {
		s.chainDiskLabel = "chain data"
		s.chainDiskPct = d.DataDiskPct
	}
	s.loadDetail = infraLoadDetail(d)
	return s
}

func infraImageShort(image string) string {
	image = strings.TrimSpace(image)
	if image == "" {
		return ""
	}
	if i := strings.LastIndex(image, "/"); i >= 0 {
		image = image[i+1:]
	}
	if i := strings.Index(image, "@sha256:"); i >= 0 {
		image = image[:i]
	}
	return image
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
		if d.DataPath != "" {
			return fmt.Sprintf("%s used of %s  (%s)", d.DataDiskUsed, d.DataDiskTotal, d.DataPath)
		}
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

func infraContainerCardHTML(d model.Report, s infraState) string {
	statusCls := "badge--ok"
	statusLabel := "running"
	if !d.NodeRunning {
		statusCls = "badge--bad"
		statusLabel = "stopped"
	}
	var b strings.Builder
	fmt.Fprintf(&b, `<div class="eco-domain eco-domain--infra">`)
	ecoDomainCardTitle(&b, "evmd-node", "Docker container", statusCls, statusLabel)
	b.WriteString(`<div class="eco-domain__rows">`)
	if s.imageShort != "" {
		ecoDomainRow(&b, "", "image", s.imageShort, "deployed container image")
	}
	if d.NodeCPU != "" && d.NodeCPU != "0.0%" {
		ecoDomainRow(&b, "", "cpu", d.NodeCPU, "container CPU vs host")
	}
	memVal := fmt.Sprintf("%s / %s", d.NodeMemUsed, d.NodeMemTotal)
	if d.NodeMemPct > 0 {
		memVal = fmt.Sprintf("%s  (%d%% of limit)", memVal, d.NodeMemPct)
	}
	ecoDomainRow(&b, infraContainerRowClass(d.NodeMemPct), "memory", memVal, "cgroup memory limit")
	if d.NodeUptime != "" {
		ecoDomainRow(&b, "", "uptime", d.NodeUptime, "since last start")
	}
	if d.NodeStartedAt != "" {
		ecoDomainRow(&b, "", "started at", d.NodeStartedAt, "container start timestamp")
	}
	if d.Restarts > 0 {
		ecoDomainRow(&b, infraRestartRowClass(d.Restarts), "restarts", fmt.Sprintf("%d", d.Restarts), "Docker restart count")
	}
	if d.NodeOOMKilled {
		ecoDomainRow(&b, "eco-domain__row--warn", "oom killed", "yes", "last exit was OOM")
	}
	if d.DataPath != "" {
		ecoDomainRow(&b, "", "data path", d.DataPath, "validator home on host")
	}
	b.WriteString(`</div></div>`)
	return b.String()
}

func infraContainerRowClass(memPct int) string {
	switch infraMeterTone(memPct) {
	case "bad":
		return "eco-domain__row--warn"
	case "warn":
		return "eco-domain__row--warn"
	default:
		return ""
	}
}

func infraRestartRowClass(n int) string {
	if n >= 3 {
		return "eco-domain__row--warn"
	}
	return ""
}

func infraSummaryFootHTML(d model.Report) string {
	var parts []string
	if d.NodeUptime != "" {
		parts = append(parts, fmt.Sprintf("up <strong>%s</strong>", html.EscapeString(d.NodeUptime)))
	}
	if d.NodeCPU != "" && d.NodeCPU != "0.0%" {
		parts = append(parts, fmt.Sprintf("CPU <strong>%s</strong>", html.EscapeString(d.NodeCPU)))
	}
	if d.NodeMemPct > 0 {
		parts = append(parts, fmt.Sprintf("container RAM <strong>%d%%</strong>", d.NodeMemPct))
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, " · ")
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
