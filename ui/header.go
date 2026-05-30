package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func renderHeader(s *unifiedSnapshot, width int) string {
	if s == nil {
		return headerStyle.Width(width - 2).Render(styleBold.Render("pmtop") + "  " + dim("fetching…"))
	}

	chain := s.Chain
	sys := s.System
	docker := s.Docker

	// sync status
	var syncStr string
	if chain.Err != nil {
		syncStr = bad("RPC unavailable")
	} else if chain.CatchingUp {
		syncStr = bad("CATCHING UP ✗")
	} else {
		syncStr = ok("SYNCED ✓")
	}

	// block info
	heightStr := fmt.Sprintf("%-12d", chain.BlockHeight)
	var intervalStr string
	if chain.BlockInterval > 0 {
		intervalStr = fmt.Sprintf("%.2fs/blk", chain.BlockInterval.Seconds())
	} else {
		intervalStr = "?s/blk"
	}

	// peers
	peerStr := fmt.Sprintf("%d peers", chain.PeerCount)

	// container
	var containerStr string
	if docker.Err != nil {
		containerStr = bad("container unavailable")
	} else if docker.Running {
		containerStr = ok("container running ✓")
	} else {
		containerStr = bad("container stopped ✗")
	}

	// CPU (docker container CPU)
	cpuStr := fmt.Sprintf("CPU %.1f%%", docker.CPUPercent)

	// RAM
	var ramStr string
	if sys.MemTotal > 0 {
		ramPct := pct(sys.MemTotal-sys.MemAvail, sys.MemTotal)
		ramStr = fmt.Sprintf("RAM %.0f%%", ramPct)
	} else {
		ramStr = "RAM ?%"
	}

	// Disk
	var diskStr string
	if sys.DiskTotal > 0 {
		diskPct := pct(sys.DiskUsed, sys.DiskTotal)
		raw := fmt.Sprintf("Disk %.0f%%", diskPct)
		diskStr = colorDiskLine(raw, diskPct)
	} else {
		diskStr = "Disk ?%"
	}

	// time
	timeStr := time.Now().UTC().Format("15:04:05") + " UTC"

	moniker := chain.Moniker
	if moniker == "" {
		moniker = "unknown"
	}
	nodeID := chain.NodeID
	if len(nodeID) > 8 {
		nodeID = nodeID[:8]
	}

	title := styleBold.Render("pmtop") + " · " + moniker + " · " + nodeID

	line1 := lipgloss.JoinHorizontal(lipgloss.Top,
		title,
		lipgloss.NewStyle().Width(width-lipgloss.Width(title)-lipgloss.Width(timeStr)-4).Render(""),
		dim(timeStr),
	)
	line2 := fmt.Sprintf("  %s  ·  %s  ·  %s  ·  %s", heightStr, intervalStr, syncStr, peerStr)
	line3 := fmt.Sprintf("  %s  ·  %s  ·  %s  ·  %s", containerStr, cpuStr, ramStr, diskStr)

	return headerStyle.Width(width - 2).Render(line1 + "\n" + line2 + "\n" + line3)
}
