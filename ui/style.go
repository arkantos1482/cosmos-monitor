package ui

import "github.com/charmbracelet/lipgloss"

var (
	colorGreen  = lipgloss.Color("10")
	colorRed    = lipgloss.Color("9")
	colorYellow = lipgloss.Color("11")
	colorDim    = lipgloss.Color("240")
	colorBlue   = lipgloss.Color("12")
	colorWhite  = lipgloss.Color("15")

	styleOK      = lipgloss.NewStyle().Foreground(colorGreen)
	styleErr     = lipgloss.NewStyle().Foreground(colorRed)
	styleWarn    = lipgloss.NewStyle().Foreground(colorYellow)
	styleDim     = lipgloss.NewStyle().Foreground(colorDim)
	styleLabel   = lipgloss.NewStyle().Foreground(colorBlue)
	styleBold    = lipgloss.NewStyle().Bold(true)
	styleStrike  = lipgloss.NewStyle().Foreground(colorRed).Strikethrough(true)

	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(colorDim).
			Padding(0, 1)

	headerStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(colorDim).
			Padding(0, 1).
			Bold(true)
)

func ok(s string) string    { return styleOK.Render(s) }
func bad(s string) string   { return styleErr.Render(s) }
func warn(s string) string  { return styleWarn.Render(s) }
func dim(s string) string   { return styleDim.Render(s) }
func label(s string) string { return styleLabel.Render(s) }

func pct(used, total uint64) float64 {
	if total == 0 {
		return 0
	}
	return float64(used) / float64(total) * 100
}

func bar(ratio float64, width int) string {
	filled := int(ratio * float64(width))
	if filled > width {
		filled = width
	}
	empty := width - filled
	b := ""
	for i := 0; i < filled; i++ {
		b += "█"
	}
	for i := 0; i < empty; i++ {
		b += "░"
	}
	return b
}

func diskColor(pct float64) string {
	switch {
	case pct > 95:
		return "9"
	case pct > 85:
		return "11"
	default:
		return ""
	}
}

func colorDiskLine(line string, pct float64) string {
	c := diskColor(pct)
	if c == "" {
		return line
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color(c)).Render(line)
}
