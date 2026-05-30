package ui

import (
	"fmt"
	"time"
)

func renderSystem(s *unifiedSnapshot, width, height int) string {
	if s == nil {
		return panelStyle.Width(width).Height(height).Render(dim("fetching…"))
	}

	sys := s.System
	docker := s.Docker

	title := styleBold.Render("SYSTEM")
	if sys.Err != nil {
		return panelStyle.Width(width).Height(height).Render(title + "\n\n" + dim("unavailable: "+sys.Err.Error()))
	}

	lines := []string{title, ""}

	// CPU load
	lines = append(lines, fmt.Sprintf("%-10s %.2f  %.2f  %.2f", label("Load"), sys.LoadAvg1, sys.LoadAvg5, sys.LoadAvg15))

	// RAM
	used := sys.MemTotal - sys.MemAvail
	ramPct := pct(used, sys.MemTotal)
	lines = append(lines, fmt.Sprintf("%-10s %s / %s  %.0f%%",
		label("RAM"),
		fmtBytes(used), fmtBytes(sys.MemTotal),
		ramPct))

	// Disk
	diskPct := pct(sys.DiskUsed, sys.DiskTotal)
	diskLine := fmt.Sprintf("%-10s %s / %s  %.0f%%",
		label("Disk"),
		fmtBytes(sys.DiskUsed), fmtBytes(sys.DiskTotal),
		diskPct)
	lines = append(lines, colorDiskLine(diskLine, diskPct))
	lines = append(lines, "")

	// Docker section
	lines = append(lines, fmt.Sprintf("%-10s", styleBold.Render("Container")))

	if docker.Err != nil {
		lines = append(lines, dim("  unavailable: "+docker.Err.Error()))
	} else {
		var statusStr string
		if docker.Running {
			statusStr = ok("running ✓")
		} else {
			statusStr = bad("stopped ✗")
		}
		lines = append(lines, fmt.Sprintf("  %-9s %s", label("Status"), statusStr))
		lines = append(lines, fmt.Sprintf("  %-9s %d", label("Restarts"), docker.RestartCount))

		if docker.Running && !docker.StartedAt.IsZero() {
			uptime := time.Since(docker.StartedAt)
			lines = append(lines, fmt.Sprintf("  %-9s %s", label("Uptime"), fmtDuration(uptime)))
		}

		containerPct := docker.CPUPercent
		lines = append(lines, fmt.Sprintf("  %-9s %.1f%%", label("CPU"), containerPct))

		if docker.MemLimit > 0 {
			lines = append(lines, fmt.Sprintf("  %-9s %s / %s",
				label("RAM"),
				fmtBytes(docker.MemUsage), fmtBytes(docker.MemLimit)))
		}
	}

	content := ""
	for i, l := range lines {
		if i > 0 {
			content += "\n"
		}
		content += l
	}
	return panelStyle.Width(width).Height(height).Render(content)
}

func fmtBytes(b uint64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)
	switch {
	case b >= GB:
		return fmt.Sprintf("%.1f GB", float64(b)/GB)
	case b >= MB:
		return fmt.Sprintf("%.1f MB", float64(b)/MB)
	case b >= KB:
		return fmt.Sprintf("%.1f KB", float64(b)/KB)
	default:
		return fmt.Sprintf("%d B", b)
	}
}

func fmtDuration(d time.Duration) string {
	d = d.Round(time.Second)
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	mins := int(d.Minutes()) % 60
	if days > 0 {
		return fmt.Sprintf("%dd %dh", days, hours)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, mins)
	}
	return fmt.Sprintf("%dm", mins)
}
