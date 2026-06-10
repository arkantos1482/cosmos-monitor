package panel

import (
	"fmt"
	"html"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func writeInfraSummary(w Writer, d model.Report, mode SummaryMode) {
	nodeStatus := "stopped"
	nodeKind := "bad"
	if d.NodeRunning {
		nodeStatus = "running"
		nodeKind = "ok"
	}

	summaryWrapStart(w, mode, "infra")
	w.WriteHTML(`<div class="infra-summary">`)
	w.WriteHTML(`<div class="infra-summary__gauges">`)
	writeMiniGauge(w, "RAM", d.MemPct)
	writeMiniGauge(w, "Disk", d.DiskPct)
	writeMiniGauge(w, "Load 1m", loadGaugePct(d.Load1))
	w.WriteHTML(`</div>`)
	w.WriteHTML(`<div class="infra-summary__container">`)
	writeSummaryBadges(w, "infra-summary__status", summaryBadge{nodeStatus, nodeKind})
	w.WriteHTML(fmt.Sprintf(`<p class="infra-summary__row">CPU <strong>%s</strong> · RAM <strong>%s / %s</strong></p>`,
		html.EscapeString(d.NodeCPU), html.EscapeString(d.NodeMemUsed), html.EscapeString(d.NodeMemTotal)))
	w.WriteHTML(fmt.Sprintf(`<p class="infra-summary__row">Restarts <strong>%d</strong> · uptime <strong>%s</strong></p>`,
		d.Restarts, html.EscapeString(orDash(d.NodeUptime))))
	w.WriteHTML(fmt.Sprintf(`<p class="infra-summary__load mono">load %.2f / %.2f / %.2f (1m 5m 15m)</p>`,
		d.Load1, d.Load5, d.Load15))
	w.WriteHTML(`</div></div>`)
	summaryWrapEnd(w, mode)
}

func writeInfra(w Writer, d model.Report) {
	writeInfraSummary(w, d, SummaryEmbedded)
	w.Section("1. INFRASTRUCTURE")
	w.Em("Host and container for this node.")

	w.Subsection("OS")
	w.Hint("`load` → proc /proc/loadavg; `ram` → proc /proc/meminfo (MemTotal, MemAvailable); `disk` → fs statfs /.")
	w.Row("load", fmt.Sprintf("%.2f / %.2f / %.2f  (1m 5m 15m)", d.Load1, d.Load5, d.Load15))
	w.Row("ram", fmt.Sprintf("%s / %s  (%d%%)", d.MemUsed, d.MemTotal, d.MemPct))
	w.Row("disk", fmt.Sprintf("%s / %s  (%d%%)", d.DiskUsed, d.DiskTotal, d.DiskPct))

	w.Subsection("Container")
	w.Hint("`status`, `restarts`, `uptime` → docker GET /containers/{name}/json; `cpu`, `ram` → docker GET /containers/{name}/stats?stream=false (unix socket).")
	nodeStatus := "stopped"
	if d.NodeRunning {
		nodeStatus = "running"
	}
	w.Row("status", nodeStatus)
	w.Row("cpu", d.NodeCPU)
	w.Row("ram", fmt.Sprintf("%s / %s", d.NodeMemUsed, d.NodeMemTotal))
	w.Row("restarts", fmt.Sprintf("%d", d.Restarts))
	if d.NodeUptime != "" {
		w.Row("uptime", d.NodeUptime)
	}
}
