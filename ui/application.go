package ui

import (
	"fmt"

	"github.com/arkantos1482/cosmos-monitor/fetch"
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
	supplyStr := fetch.FormatCoin(chain.TotalSupply, chain.TotalSupplyDenom)
	lines = append(lines, fmt.Sprintf("%-12s %s", label("Supply"), supplyStr))

	// Bonded — convert both to display units for the percentage calculation
	bondedF, _ := fetch.NormalizeCoin(chain.BondedTokens, chain.TotalSupplyDenom)
	totalF, _ := fetch.NormalizeCoin(chain.TotalSupply, chain.TotalSupplyDenom)
	var bondedPct float64
	if totalF > 0 {
		bondedPct = bondedF / totalF * 100
	}
	bondedStr := fetch.FormatCoin(chain.BondedTokens, chain.TotalSupplyDenom)
	lines = append(lines, fmt.Sprintf("%-12s %s  %.1f%%", label("Bonded"), bondedStr, bondedPct))

	// bonded bar
	lines = append(lines, fmt.Sprintf("%-12s %s", "", bar(bondedPct/100, 12)))

	// Inflation
	lines = append(lines, fmt.Sprintf("%-12s %.2f%%", label("Inflation"), chain.Inflation*100))

	// Community pool
	lines = append(lines, fmt.Sprintf("%-12s %s", label("Community"), chain.CommunityPool))

	// Params
	p := chain.Params
	lines = append(lines, fmt.Sprintf("%-12s %.2f%%", label("Comm. tax"), p.CommunityTax*100))
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
