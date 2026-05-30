package ui

import (
	"fmt"
	"strconv"
)

func renderApplication(s *unifiedSnapshot, width, height int) string {
	if s == nil {
		return panelStyle.Width(width).Height(height).Render(dim("fetching…"))
	}

	chain := s.Chain
	title := styleBold.Render("APPLICATION")
	if chain.Err != nil {
		return panelStyle.Width(width).Height(height).Render(title + "\n\n" + dim("unavailable: "+chain.Err.Error()))
	}

	lines := []string{title, ""}

	// Supply
	supplyStr := formatTokenAmount(chain.TotalSupply, chain.TotalSupplyDenom)
	lines = append(lines, fmt.Sprintf("%-12s %s", label("Supply"), supplyStr))

	// Bonded
	bondedF, _ := strconv.ParseFloat(chain.BondedTokens, 64)
	totalF, _ := strconv.ParseFloat(chain.TotalSupply, 64)
	var bondedPct float64
	if totalF > 0 {
		bondedPct = bondedF / totalF * 100
	}
	bondedStr := formatTokenAmount(chain.BondedTokens, chain.TotalSupplyDenom)
	lines = append(lines, fmt.Sprintf("%-12s %s  %.1f%%", label("Bonded"), bondedStr, bondedPct))

	// bonded bar
	barStr := bar(bondedPct/100, 12)
	lines = append(lines, fmt.Sprintf("%-12s %s", "", barStr))

	// Inflation
	lines = append(lines, fmt.Sprintf("%-12s %.2f%%", label("Inflation"), chain.Inflation*100))

	// Community pool
	lines = append(lines, fmt.Sprintf("%-12s %s", label("Community"), chain.CommunityPool))

	// Params
	p := chain.Params
	lines = append(lines, fmt.Sprintf("%-12s %.0f%%", label("Goal bond"), p.GoalBonded*100))

	// Unbonding time
	unbondH := int(p.UnbondingTime.Hours())
	if unbondH > 0 {
		lines = append(lines, fmt.Sprintf("%-12s %dh", label("Unbonding"), unbondH))
	} else {
		lines = append(lines, fmt.Sprintf("%-12s %s", label("Unbonding"), dim("?")))
	}

	content := joinLines(lines)
	return panelStyle.Width(width).Height(height).Render(content)
}

func formatTokenAmount(amount, denom string) string {
	f, err := strconv.ParseFloat(amount, 64)
	if err != nil || amount == "" {
		return dim("?")
	}
	var s string
	switch {
	case f >= 1e12:
		s = fmt.Sprintf("%.2fT", f/1e12)
	case f >= 1e9:
		s = fmt.Sprintf("%.2fB", f/1e9)
	case f >= 1e6:
		s = fmt.Sprintf("%.2fM", f/1e6)
	case f >= 1e3:
		s = fmt.Sprintf("%.2fK", f/1e3)
	default:
		s = fmt.Sprintf("%.4f", f)
	}
	if denom != "" {
		s += " " + denom
	}
	return s
}
