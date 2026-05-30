package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func renderGovernance(s *unifiedSnapshot, width int) string {
	title := styleBold.Render("GOVERNANCE")
	if s == nil {
		return panelStyle.Width(width - 2).Render(title + "  " + dim("fetching…"))
	}

	chain := s.Chain
	if chain.Err != nil {
		return panelStyle.Width(width - 2).Render(title + "  " + dim("unavailable"))
	}

	govParts := []string{title}
	if len(chain.Proposals) == 0 {
		govParts = append(govParts, dim("  no active proposals"))
	} else {
		for _, p := range chain.Proposals {
			var countdown string
			switch p.Status {
			case "PROPOSAL_STATUS_VOTING_PERIOD":
				remain := time.Until(p.VotingEnd)
				countdown = fmt.Sprintf("VOTING   ends %s", fmtRemaining(remain))
			case "PROPOSAL_STATUS_DEPOSIT_PERIOD":
				remain := time.Until(p.DepositEnd)
				countdown = fmt.Sprintf("DEPOSIT  ends %s", fmtRemaining(remain))
			default:
				countdown = p.Status
			}
			govParts = append(govParts, fmt.Sprintf("  #%-4d %-30s %s",
				p.ID, truncate(p.Title, 30), countdown))
		}
	}

	// Upgrade
	upgradeParts := []string{styleBold.Render("UPGRADE")}
	if chain.UpgradeName != "" {
		upgradeStr := fmt.Sprintf("  %s at height %s", chain.UpgradeName, fmtInt64(chain.UpgradeHeight))
		upgradeParts = append(upgradeParts, warn(upgradeStr))
	} else {
		upgradeParts = append(upgradeParts, dim("  none pending"))
	}

	govSection := strings.Join(govParts, "\n")
	upgradeSection := strings.Join(upgradeParts, "\n")

	// side by side within the strip
	halfW := (width - 6) / 2
	left := lipgloss.NewStyle().Width(halfW).Render(govSection)
	right := lipgloss.NewStyle().Width(halfW).Render(upgradeSection)
	content := lipgloss.JoinHorizontal(lipgloss.Top, left, "  ", right)

	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(colorDim).
		Padding(0, 1).
		Width(width - 2).
		Render(content)
}

func fmtRemaining(d time.Duration) string {
	if d < 0 {
		return "expired"
	}
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	if days > 0 {
		return fmt.Sprintf("%dd %dh", days, hours)
	}
	mins := int(d.Minutes()) % 60
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, mins)
	}
	return fmt.Sprintf("%dm", mins)
}
