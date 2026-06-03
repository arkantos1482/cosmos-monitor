package main

import (
	"fmt"
	"io"
	"time"

	"github.com/arkantos1482/cosmos-monitor/fetch"
	"github.com/charmbracelet/glamour"
)

func printDashboard(w io.Writer, chain fetch.ChainSnapshot, ev fetch.EVMSnapshot, sys fetch.SystemSnapshot, docker fetch.DockerSnapshot) {
	d := buildWebData(chain, ev, sys, docker)
	md := buildMarkdown(d, false)
	r, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap(0),
	)
	if err != nil {
		fmt.Fprint(w, md)
		return
	}
	rendered, err := r.Render(md)
	if err != nil {
		fmt.Fprint(w, md)
		return
	}
	fmt.Fprint(w, rendered)
}

// ── helpers ──────────────────────────────────────────────────────────────────

func fmtInt(n int64) string {
	if n <= 0 {
		return fmt.Sprintf("%d", n)
	}
	s := fmt.Sprintf("%d", n)
	out := make([]byte, 0, len(s)+4)
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, byte(c))
	}
	return string(out)
}

func fmtBytes(b uint64) string {
	const (KB = 1024; MB = 1024 * KB; GB = 1024 * MB)
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

func fmtDur(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}

func fmtDurFull(d time.Duration) string {
	if d == 0 {
		return "0"
	}
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

func boolStr(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}

func fmtFraction(s string) string {
	v := 0.0
	fmt.Sscanf(s, "%f", &v)
	return fmt.Sprintf("%.4f%%", v*100)
}
