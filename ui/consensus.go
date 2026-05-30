package ui

import (
	"fmt"
	"sort"
)

func renderConsensus(s *unifiedSnapshot, width, height int) string {
	if s == nil {
		return panelStyle.Width(width).Height(height).Render(dim("fetching…"))
	}

	chain := s.Chain
	title := styleBold.Render("CONSENSUS")
	if chain.Err != nil {
		return panelStyle.Width(width).Height(height).Render(title + "\n\n" + dim("unavailable: "+chain.Err.Error()))
	}

	lines := []string{title, ""}

	lines = append(lines, fmt.Sprintf("%-12s %s", label("Height"), fmtInt64(chain.BlockHeight)))

	if chain.BlockInterval > 0 {
		lines = append(lines, fmt.Sprintf("%-12s %.2fs", label("Block"), chain.BlockInterval.Seconds()))
	} else {
		lines = append(lines, fmt.Sprintf("%-12s %s", label("Block"), dim("?")))
	}

	var syncStr string
	if chain.CatchingUp {
		syncStr = bad("CATCHING UP ✗")
	} else {
		syncStr = ok("SYNCED ✓")
	}
	lines = append(lines, fmt.Sprintf("%-12s %s", label("Sync"), syncStr))
	lines = append(lines, fmt.Sprintf("%-12s %d", label("Peers"), chain.PeerCount))
	lines = append(lines, "")

	// Validator counts
	bonded, jailed, tombstoned := 0, 0, 0
	for _, v := range chain.Validators {
		if v.Status == "BONDED" {
			bonded++
		}
		if v.Jailed {
			jailed++
		}
		if v.Tombstoned {
			tombstoned++
		}
	}
	lines = append(lines, styleBold.Render("Validators"))
	lines = append(lines, fmt.Sprintf("  %d bonded · %d jailed · %d tombstoned",
		bonded, jailed, tombstoned))

	// Next proposer (highest proposer priority)
	if len(chain.Validators) > 0 {
		sorted := make([]struct {
			moniker  string
			priority int64
		}, 0, len(chain.Validators))
		for _, v := range chain.Validators {
			sorted = append(sorted, struct {
				moniker  string
				priority int64
			}{v.Moniker, v.ProposerPriority})
		}
		sort.Slice(sorted, func(i, j int) bool { return sorted[i].priority > sorted[j].priority })
		lines = append(lines, fmt.Sprintf("  next: %s", sorted[0].moniker))
	}
	lines = append(lines, "")

	// Slashing params
	p := chain.Params
	lines = append(lines, fmt.Sprintf("%-12s %s blocks", label("Sign window"), fmtInt64(p.SignedBlocksWindow)))
	lines = append(lines, fmt.Sprintf("%-12s %.0f%%", label("Min signed"), p.MinSignedPerWindow*100))

	content := joinLines(lines)
	return panelStyle.Width(width).Height(height).Render(content)
}

func fmtInt64(n int64) string {
	// add thousands separators
	s := fmt.Sprintf("%d", n)
	out := ""
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			out += ","
		}
		out += string(c)
	}
	return out
}

func joinLines(lines []string) string {
	s := ""
	for i, l := range lines {
		if i > 0 {
			s += "\n"
		}
		s += l
	}
	return s
}
