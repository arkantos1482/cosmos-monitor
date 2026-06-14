package panel

import (
	"fmt"
	"html"
	"strings"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func writeInfraSummary(w Writer, d model.Report, mode SummaryMode) {
	nodeStatus := "stopped"
	nodeKind := "bad"
	if d.NodeRunning {
		nodeStatus = "running"
		nodeKind = "ok"
	}

	diskLabel := "disk"
	diskPct := d.DiskPct
	if d.DataPath != "" && d.DataDiskPct > 0 {
		diskLabel = "chain data"
		diskPct = d.DataDiskPct
	}

	summaryWrapStart(w, mode, "infra")
	w.WriteHTML(`<div class="infra-summary">`)
	w.WriteHTML(`<div class="infra-summary__top">`)
	writeSummaryBadges(w, "infra-summary__status", summaryBadge{nodeStatus, nodeKind})
	if img := infraImageLabel(d.NodeImage); img != "" {
		w.WriteHTML(fmt.Sprintf(`<span class="infra-summary__hint">%s</span>`, html.EscapeString(img)))
	}
	w.WriteHTML(`</div>`)
	w.WriteHTML(`<div class="infra-summary__gauges">`)
	writeMiniGauge(w, "RAM", d.MemPct)
	writeMiniGauge(w, diskLabel, diskPct)
	writeMiniGauge(w, "load 1m", loadGaugePct(d.Load1))
	w.WriteHTML(`</div>`)
	if line := infraContainerHeadline(d); line != "" {
		w.WriteHTML(fmt.Sprintf(`<p class="infra-summary__row">%s</p>`, line))
	}
	w.WriteHTML(`</div>`)
	summaryWrapEnd(w, mode)
}

func infraContainerHeadline(d model.Report) string {
	var parts []string
	if d.NodeCPU != "" && d.NodeCPU != "0.0%" {
		parts = append(parts, fmt.Sprintf("CPU <strong>%s</strong>", html.EscapeString(d.NodeCPU)))
	}
	if d.NodeMemPct > 0 {
		parts = append(parts, fmt.Sprintf("RAM <strong>%d%%</strong> of limit", d.NodeMemPct))
	} else if d.NodeMemUsed != "" && d.NodeMemTotal != "" {
		parts = append(parts, fmt.Sprintf("RAM <strong>%s / %s</strong>",
			html.EscapeString(d.NodeMemUsed), html.EscapeString(d.NodeMemTotal)))
	}
	if d.NodeUptime != "" {
		parts = append(parts, fmt.Sprintf("up <strong>%s</strong>", html.EscapeString(d.NodeUptime)))
	}
	if d.Restarts > 0 {
		parts = append(parts, fmt.Sprintf("<strong>%d</strong> restarts", d.Restarts))
	}
	if d.NodeOOMKilled {
		parts = append(parts, `<span class="badge badge--bad">OOM killed</span>`)
	}
	return strings.Join(parts, " · ")
}

func writeInfra(w Writer, d model.Report) {
	w.Section("1. INFRASTRUCTURE")
	writeEmbeddedSectionIntro(w, "Host resource pressure and the `evmd-node` container — CPU, memory, disk, and process health.")
	writeInfraSummary(w, d, SummaryEmbedded)
	writeInfraHost(w, d)
	writeInfraContainer(w, d)
	writeSectionSources(w, ViewInfra, d)
}

func writeInfraHost(w Writer, d model.Report) {
	w.Layer("Host")

	w.Row("load", infraLoadValue(d))
	w.Row("memory", fmt.Sprintf("%s used / %s total  (%d%%) · %s free",
		d.MemUsed, d.MemTotal, d.MemPct, orDash(d.MemAvail)))

	if d.DataPath != "" {
		w.Row("chain data", fmt.Sprintf("%s / %s  (%d%%)  _(%s)_",
			d.DataDiskUsed, d.DataDiskTotal, d.DataDiskPct, d.DataPath))
	}
	w.Row("disk", fmt.Sprintf("%s used / %s total  (%d%%) · %s free",
		d.DiskUsed, d.DiskTotal, d.DiskPct, orDash(d.DiskAvail)))

	if d.SwapTotal != "" {
		w.Row("swap", fmt.Sprintf("%s / %s", orDash(d.SwapUsed), d.SwapTotal))
	}
}

func writeInfraContainer(w Writer, d model.Report) {
	w.Layer("evmd-node")

	if d.NodeImage != "" {
		w.Row("image", infraImageLabel(d.NodeImage))
	}
	w.Row("cpu", orDash(d.NodeCPU))
	if d.NodeMemPct > 0 {
		w.Row("memory", fmt.Sprintf("%s / %s  (%d%%)", d.NodeMemUsed, d.NodeMemTotal, d.NodeMemPct))
	} else {
		w.Row("memory", fmt.Sprintf("%s / %s", d.NodeMemUsed, d.NodeMemTotal))
	}
	if d.NodeUptime != "" {
		w.Row("uptime", d.NodeUptime)
	}
	if d.NodeStartedAt != "" {
		w.Row("started", d.NodeStartedAt)
	}
	if d.Restarts > 0 {
		w.Row("restarts", fmt.Sprintf("%d", d.Restarts))
	}
	if d.NodeOOMKilled {
		w.Row("oom killed", "**yes**")
	}
}

func infraLoadValue(d model.Report) string {
	v := fmt.Sprintf("%.2f / %.2f / %.2f  (1m 5m 15m)", d.Load1, d.Load5, d.Load15)
	if d.NumCPU > 0 {
		v += fmt.Sprintf("  _(%d CPUs — %.2f per core @ 1m)_", d.NumCPU, d.Load1/float64(d.NumCPU))
	}
	return v
}

func infraImageLabel(image string) string {
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
