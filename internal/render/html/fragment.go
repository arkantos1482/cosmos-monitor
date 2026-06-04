package html

import (
	"github.com/arkantos1482/cosmos-monitor/internal/model"
	"github.com/arkantos1482/cosmos-monitor/internal/panel"
)

// RenderFragment converts a Report to an HTML fragment (no document shell).
func RenderFragment(d model.Report) string {
	return panel.Build(d)
}

// RenderView renders the home overview or one section.
func RenderView(v panel.View, d model.Report) string {
	return panel.BuildView(v, d)
}
