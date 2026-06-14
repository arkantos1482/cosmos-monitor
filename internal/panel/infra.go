package panel

import (
	"fmt"
	"html"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

const infraContainerName = "evmd-node"

func writeInfraSummary(w Writer, d model.Report, mode SummaryMode) {
	nodeStatus, nodeKind := infraNodeStatus(d)

	summaryWrapStart(w, mode, "infra")
	w.WriteHTML(`<div class="infra-summary">`)
	w.WriteHTML(`<div class="infra-summary__top">`)
	w.WriteHTML(fmt.Sprintf(`<span class="infra-summary__name">%s</span>`, html.EscapeString(infraContainerName)))
	writeSummaryBadges(w, "infra-summary__badges", summaryBadge{nodeStatus, nodeKind})
	w.WriteHTML(`</div>`)

	w.WriteHTML(`<div class="infra-summary__gauges">`)
	writeMiniGauge(w, "host RAM", d.MemPct)
	writeMiniGauge(w, "host disk", d.DiskPct)
	writeMiniGauge(w, "load 1m", loadGaugePct(d.Load1))
	w.WriteHTML(`</div>`)

	w.WriteHTML(`<div class="infra-summary__kpis">`)
	writeInfraSummaryKPI(w, "host RAM", fmt.Sprintf("%s / %s", d.MemUsed, d.MemTotal))
	writeInfraSummaryKPI(w, "host disk", fmt.Sprintf("%s / %s", d.DiskUsed, d.DiskTotal))
	writeInfraSummaryKPI(w, "load avg", fmt.Sprintf("%.2f · %.2f · %.2f", d.Load1, d.Load5, d.Load15))
	writeInfraSummaryKPI(w, "container CPU", orDash(d.NodeCPU))
	writeInfraSummaryKPI(w, "container RAM", fmt.Sprintf("%s / %s", d.NodeMemUsed, d.NodeMemTotal))
	writeInfraSummaryKPI(w, "restarts", fmt.Sprintf("%d · %s uptime", d.Restarts, orDash(d.NodeUptime)))
	w.WriteHTML(`</div></div>`)
	summaryWrapEnd(w, mode)
}

func writeInfraSummaryKPI(w Writer, label, value string) {
	if value == "" || value == " / " || value == "0 · — uptime" {
		return
	}
	w.WriteHTML(fmt.Sprintf(
		`<div class="infra-summary__kpi"><span class="infra-summary__kpi-label">%s</span>`+
			`<span class="infra-summary__kpi-val">%s</span></div>`,
		html.EscapeString(label), html.EscapeString(value)))
}

func writeInfra(w Writer, d model.Report) {
	w.Section("1. INFRASTRUCTURE")
	writeEmbeddedSectionIntro(w, "Host CPU load, memory, and disk, plus the `evmd-node` Docker container cgroup usage and lifecycle.")
	writeInfraSummary(w, d, SummaryEmbedded)
	writeInfraCompare(w, d)
	writeSectionSources(w, ViewInfra, d)
}

func writeInfraCompare(w Writer, d model.Report) {
	nodeStatus, nodeKind := infraNodeStatus(d)

	w.Subsection("Host vs container")
	w.WriteHTML(`<div class="infra-compare-wrap"><table class="data-table infra-compare">`)
	w.WriteHTML(`<thead><tr><th>resource</th><th>host</th><th>` + html.EscapeString(infraContainerName) + `</th></tr></thead><tbody>`)

	writeInfraCompareRow(w, "CPU", "—", orDash(d.NodeCPU))
	writeInfraCompareRowHTML(w, "memory",
		infraHostResourceCell(d.MemUsed, d.MemTotal, d.MemPct),
		infraContainerMemCell(d.NodeMemUsed, d.NodeMemTotal))
	writeInfraCompareRowHTML(w, "disk",
		infraHostResourceCell(d.DiskUsed, d.DiskTotal, d.DiskPct),
		`<span class="id-empty">—</span>`)
	writeInfraCompareRow(w, "load avg (1m · 5m · 15m)",
		fmt.Sprintf("%.2f · %.2f · %.2f", d.Load1, d.Load5, d.Load15), "—")
	writeInfraCompareRowHTML(w, "status", "—",
		fmt.Sprintf(`<span class="badge badge--%s">%s</span>`, nodeKind, html.EscapeString(nodeStatus)))
	writeInfraCompareRow(w, "restarts", "—", fmt.Sprintf("%d", d.Restarts))
	if d.NodeUptime != "" {
		writeInfraCompareRow(w, "uptime", "—", d.NodeUptime)
	}

	w.WriteHTML(`</tbody></table></div>`)
}

func writeInfraCompareRow(w Writer, label, host, container string) {
	writeInfraCompareRowHTML(w, label, html.EscapeString(host), html.EscapeString(container))
}

func writeInfraCompareRowHTML(w Writer, label, hostHTML, containerHTML string) {
	w.WriteHTML(fmt.Sprintf(
		`<tr><td class="infra-compare__metric">%s</td><td class="infra-compare__host">%s</td><td class="infra-compare__container">%s</td></tr>`,
		html.EscapeString(label), hostHTML, containerHTML))
}

func infraHostResourceCell(used, total string, pct int) string {
	if used == "" && total == "" {
		return `<span class="id-empty">—</span>`
	}
	cell := fmt.Sprintf(`<span class="infra-compare__amount">%s / %s</span>`,
		html.EscapeString(used), html.EscapeString(total))
	if pct > 0 {
		cell += fmt.Sprintf(` <span class="infra-compare__pct">(%d%%)</span>`, pct)
		cell += infraPctBar(pct)
	}
	return cell
}

func infraContainerMemCell(used, total string) string {
	if used == "" && total == "" {
		return `<span class="id-empty">—</span>`
	}
	return fmt.Sprintf(`<span class="infra-compare__amount">%s / %s</span>`,
		html.EscapeString(used), html.EscapeString(total))
}

func infraPctBar(pct int) string {
	if pct <= 0 {
		return ""
	}
	if pct > 100 {
		pct = 100
	}
	return fmt.Sprintf(`<div class="kpi-bar infra-compare__bar"><div class="kpi-bar__fill" style="width:%d%%"></div></div>`, pct)
}

func infraNodeStatus(d model.Report) (status, kind string) {
	if d.NodeRunning {
		return "running", "ok"
	}
	return "stopped", "bad"
}
