package html

import (
	"github.com/arkantos1482/cosmos-monitor/internal/model"
	"github.com/arkantos1482/cosmos-monitor/internal/panel"
)

// RenderFragment converts a Report to an HTML fragment (no document shell).
func RenderFragment(d model.Report) string {
	return RenderFragmentWithOptions(d, panel.Options{})
}

// RenderFragmentWithOptions renders all sections with display options.
func RenderFragmentWithOptions(d model.Report, opts panel.Options) string {
	return panel.BuildWithOptions(d, opts)
}

// RenderView renders the home overview or one section.
func RenderView(v panel.View, d model.Report) string {
	return RenderViewWithOptions(v, d, panel.Options{})
}

// RenderViewWithOptions renders one view with display options.
func RenderViewWithOptions(v panel.View, d model.Report, opts panel.Options) string {
	return panel.BuildViewWithOptions(v, d, opts)
}
