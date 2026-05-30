package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func renderValidators(s *unifiedSnapshot, width int) string {
	title := styleBold.Render("VALIDATORS")
	if s == nil {
		return panelStyle.Width(width - 2).Render(title + "\n" + dim("fetching…"))
	}

	chain := s.Chain
	if chain.Err != nil {
		return panelStyle.Width(width - 2).Render(title + "\n" + dim("unavailable: "+chain.Err.Error()))
	}

	if len(chain.Validators) == 0 {
		return panelStyle.Width(width - 2).Render(title + "\n" + dim("no validators"))
	}

	// Sort by voting power desc
	vals := make([]struct {
		Moniker            string
		VotingPowerPercent float64
		Commission         float64
		MissedBlocks       int64
		OutstandingRewards string
		CommissionEarned   string
		Status             string
		Jailed             bool
		Tombstoned         bool
	}, 0, len(chain.Validators))
	for _, v := range chain.Validators {
		vals = append(vals, struct {
			Moniker            string
			VotingPowerPercent float64
			Commission         float64
			MissedBlocks       int64
			OutstandingRewards string
			CommissionEarned   string
			Status             string
			Jailed             bool
			Tombstoned         bool
		}{
			v.Moniker, v.VotingPowerPercent, v.Commission,
			v.MissedBlocks, v.OutstandingRewards, v.CommissionEarned,
			v.Status, v.Jailed, v.Tombstoned,
		})
	}
	sort.Slice(vals, func(i, j int) bool {
		return vals[i].VotingPowerPercent > vals[j].VotingPowerPercent
	})

	header := dim(fmt.Sprintf("  %-20s %6s %10s %8s  %-16s %-12s %s",
		"moniker", "vp%", "commission", "missed", "outstanding", "earned", "status"))

	rows := []string{title, header}
	for _, v := range vals {
		monikerDisplay := v.Moniker
		if v.Jailed {
			monikerDisplay += " ⚠"
		}

		rowLine := fmt.Sprintf("  %-20s %5.1f%% %9.1f%% %8d  %-16s %-12s %s",
			truncate(monikerDisplay, 20),
			v.VotingPowerPercent,
			v.Commission*100,
			v.MissedBlocks,
			truncate(v.OutstandingRewards, 16),
			truncate(v.CommissionEarned, 12),
			statusBadge(v.Status),
		)

		var rowStyled string
		switch {
		case v.Tombstoned:
			rowStyled = styleStrike.Render(rowLine)
		case v.Jailed:
			rowStyled = styleErr.Render(rowLine)
		case v.MissedBlocks > 500:
			rowStyled = styleErr.Render(rowLine)
		case v.MissedBlocks > 100:
			rowStyled = styleWarn.Render(rowLine)
		default:
			rowStyled = rowLine
		}

		rows = append(rows, rowStyled)
	}

	content := strings.Join(rows, "\n")
	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(colorDim).
		Padding(0, 1).
		Width(width - 2).
		Render(content)
}

func statusBadge(status string) string {
	switch status {
	case "BONDED":
		return ok("● BON")
	case "UNBONDING":
		return warn("● UBG")
	case "UNBONDED":
		return dim("○ UBD")
	default:
		return dim(status)
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}
