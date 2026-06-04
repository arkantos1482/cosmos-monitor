package markdown

import (
	"fmt"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func writeInfra(m *mdWriter, d model.Report) {
	// ── 1. INFRASTRUCTURE ────────────────────────────────────────────────────
	m.section("1. INFRASTRUCTURE")

	m.subsection("OS")
	m.hint("`load` → `/proc/loadavg`; `ram` → `/proc/meminfo` (MemTotal, MemAvailable); `disk` → `statfs` on `/`.")
	m.row("load", fmt.Sprintf("%.2f / %.2f / %.2f  (1m 5m 15m)", d.Load1, d.Load5, d.Load15))
	m.row("ram", fmt.Sprintf("%s / %s  (%d%%)", d.MemUsed, d.MemTotal, d.MemPct))
	m.row("disk", fmt.Sprintf("%s / %s  (%d%%)", d.DiskUsed, d.DiskTotal, d.DiskPct))

	m.subsection("Container")
	m.hint("`status` / `restarts` / `uptime` → Docker `GET /containers/{name}/json`; `cpu` / `ram` → `GET /containers/{name}/stats?stream=false` (unix socket).")
	nodeStatus := "stopped"
	if d.NodeRunning {
		nodeStatus = "running"
	}
	m.row("status", nodeStatus)
	m.row("cpu", d.NodeCPU)
	m.row("ram", fmt.Sprintf("%s / %s", d.NodeMemUsed, d.NodeMemTotal))
	m.row("restarts", fmt.Sprintf("%d", d.Restarts))
	if d.NodeUptime != "" {
		m.row("uptime", d.NodeUptime)
	}
}
