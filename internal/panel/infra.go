package panel

import (
	"fmt"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func writeInfra(w Writer, d model.Report) {
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
