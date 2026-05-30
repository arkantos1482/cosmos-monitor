package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// View assembles the full TUI layout.
func (m Model) View() string {
	w := m.width
	if w == 0 {
		w = 120
	}
	h := m.height
	if h == 0 {
		h = 40
	}

	s := m.toSnapshot()

	header := renderHeader(s, w)

	// 2×2 grid
	gridW := w - 2 // account for outer margins
	halfW := gridW / 2
	panelH := 16

	systemPanel := renderSystem(s, halfW-2, panelH)
	consensusPanel := renderConsensus(s, halfW-2, panelH)
	applicationPanel := renderApplication(s, halfW-2, panelH)
	evmPanel := renderEVM(s, halfW-2, panelH)

	topRow := lipgloss.JoinHorizontal(lipgloss.Top, systemPanel, consensusPanel)
	bottomRow := lipgloss.JoinHorizontal(lipgloss.Top, applicationPanel, evmPanel)
	grid := lipgloss.JoinVertical(lipgloss.Left, topRow, bottomRow)

	validatorTable := renderValidators(s, w)
	govStrip := renderGovernance(s, w)

	var loadingBar string
	if m.loading {
		loadingBar = styleDim.Render("  fetching…") + "\n"
	}

	hint := styleDim.Render(" r:refresh  q:quit")

	_ = h // height used for future scroll support

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		grid,
		validatorTable,
		govStrip,
		loadingBar+hint,
	)
}
