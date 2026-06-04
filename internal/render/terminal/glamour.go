package terminal

import (
	"fmt"
	"io"

	"github.com/arkantos1482/cosmos-monitor/internal/markdown"
	"github.com/arkantos1482/cosmos-monitor/internal/model"
	"github.com/charmbracelet/glamour"
)

// Glamour renders GFM Markdown with ANSI styles (no Mermaid/KaTeX).
type Glamour struct {
	W     io.Writer
	Width uint
	Style string // dark, light, notty
}

func (g Glamour) Render(d model.Report) error {
	src := markdown.Build(d)
	style := g.Style
	if style == "" {
		style = "dark"
	}
	width := g.Width
	if width == 0 {
		width = 120
	}
	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(int(width)),
		glamour.WithStandardStyle(style),
	)
	if err != nil {
		return err
	}
	out, err := r.Render(src)
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(g.W, out)
	return err
}
