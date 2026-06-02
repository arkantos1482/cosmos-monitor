package main

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/arkantos1482/cosmos-monitor/fetch"
	"github.com/charmbracelet/glamour"
)

const (
	ansiReset  = "\033[0m"
	ansiBold   = "\033[1m"
	ansiDim    = "\033[2m"
	ansiRed    = "\033[31m"
	ansiGreen  = "\033[32m"
	ansiYellow = "\033[33m"
	ansiCyan   = "\033[36m"
	ansiWhite  = "\033[97m"
)

func clr(code, s string) string { return code + s + ansiReset }

// runwaySuffix returns a colored " (~N days left)" suffix for terminal display.
func runwaySuffix(d WebData) string {
	if d.PMTRunway == "" {
		return ""
	}
	if d.PMTRunwayLow {
		return fmt.Sprintf("  (%s%s%s)", ansiRed, d.PMTRunway, ansiReset)
	}
	return "  (" + d.PMTRunway + ")"
}

// pmtStatusLine builds the colored PMT rewards status string for the terminal.
func pmtStatusLine(d WebData) string {
	if !d.PMTEnabled {
		return clr(ansiDim, "disabled")
	}
	if d.PMTPoolEmpty {
		return clr(ansiRed, "ENABLED — pool EMPTY  (validators receive no PMT rewards)")
	}
	return clr(ansiGreen, "●") + " distributing  " + d.PMTRate + "   pool " + d.PMTBalance + runwaySuffix(d)
}

func printEssentials(w io.Writer, d WebData) {
	sec := func(name, summary string) {
		if summary != "" {
			fmt.Fprintf(w, "\n%s %s   %s\n", clr(ansiDim, "──"), clr(ansiBold+ansiCyan, name), summary)
		} else {
			fmt.Fprintf(w, "\n%s %s\n", clr(ansiDim, "──"), clr(ansiBold+ansiCyan, name))
		}
	}
	line := func(parts ...string) { fmt.Fprintf(w, "  %s\n", strings.Join(parts, "   ")) }
	kv   := func(k, v string) string { return clr(ansiDim, k) + " " + v }

	// ── header
	syncStr := clr(ansiGreen, "● synced")
	if !d.Synced {
		syncStr = clr(ansiRed, "● CATCHING UP")
	}
	fmt.Fprintf(w, "%s  %s  %s  %s UTC\n",
		clr(ansiBold+ansiWhite, "pmtop  "+d.Moniker),
		syncStr, d.BlockHeight, time.Now().UTC().Format("15:04:05"))

	// ── SYSTEM
	sec("SYSTEM", "")
	nodeStatus := clr(ansiRed, "stopped")
	if d.NodeRunning {
		nodeStatus = clr(ansiGreen, "● running")
	}
	uptimeStr := ""
	if d.NodeUptime != "" {
		uptimeStr = kv("up", d.NodeUptime)
	}
	line(nodeStatus,
		kv("cpu", d.NodeCPU),
		kv("ram", fmt.Sprintf("%d%%", d.MemPct)),
		kv("disk", fmt.Sprintf("%d%%", d.DiskPct)),
		kv("restarts", fmt.Sprintf("%d", d.Restarts)),
		uptimeStr)
	line(kv("load", fmt.Sprintf("%.2f/%.2f/%.2f", d.Load1, d.Load5, d.Load15)),
		kv("ram", fmt.Sprintf("%s/%s", d.MemUsed, d.MemTotal)),
		kv("disk", fmt.Sprintf("%s/%s", d.DiskUsed, d.DiskTotal)),
		kv("height", d.BlockHeight),
		kv("interval", d.BlockInterval),
		kv("peers", fmt.Sprintf("%d/%d", d.PeerCount, d.EVMPeerCount)))

	// ── VALIDATORS
	bondedCount, jailedCount := 0, 0
	for _, v := range d.Validators {
		if v.Status == "bonded" {
			bondedCount++
		}
		if v.Jailed {
			jailedCount++
		}
	}
	totalVals := len(d.Validators)
	bondedStr := clr(ansiGreen, fmt.Sprintf("%d/%d bonded", bondedCount, totalVals))
	if bondedCount < totalVals {
		bondedStr = clr(ansiRed, fmt.Sprintf("%d/%d bonded", bondedCount, totalVals))
	}
	jailedStr := fmt.Sprintf("%d jailed", jailedCount)
	if jailedCount > 0 {
		jailedStr = clr(ansiRed, jailedStr)
	}
	belowStr := fmt.Sprintf("%d below threshold", d.BelowThreshold)
	if d.BelowThreshold > 0 {
		belowStr = clr(ansiYellow, belowStr)
	}
	sec("VALIDATORS", bondedStr+"   "+jailedStr+"   "+belowStr)

	for _, v := range d.Validators {
		flags := ""
		if v.Jailed {
			flags += "  " + clr(ansiRed, "JAILED")
		}
		if v.Tombstoned {
			flags += "  " + clr(ansiRed, "TOMBSTONED")
		}
		missedStr := fmt.Sprintf("%d missed", v.Missed)
		if v.MissedHigh {
			missedStr = clr(ansiRed, missedStr)
		}
		earned := v.Outstanding
		if earned == "" {
			earned = "–"
		}
		fmt.Fprintf(w, "  %-16s  %5.1f%%  %-14s  %s  earned %s%s\n",
			truncate(v.Moniker, 16), v.VPFloat,
			missedStr, v.Status, earned, flags)
	}

	// ── ECONOMICS
	sec("ECONOMICS", "")
	line(kv("supply", d.TotalSupply),
		kv("bonded", fmt.Sprintf("%s  %.1f%%  (goal %.0f%%)", d.BondedAmt, d.BondedPct, d.GoalBonded)))
	line(kv("rewards", pmtStatusLine(d)))
	fmt.Fprintln(w)
}

func printDashboard(w io.Writer, chain fetch.ChainSnapshot, ev fetch.EVMSnapshot, sys fetch.SystemSnapshot, docker fetch.DockerSnapshot) {
	d := buildWebData(chain, ev, sys, docker)
	printEssentials(w, d)
	md := buildMarkdown(d)
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
