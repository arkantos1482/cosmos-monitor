package html

import (
	"bytes"
	"html"

	"github.com/arkantos1482/cosmos-monitor/internal/markdown"
	"github.com/arkantos1482/cosmos-monitor/internal/model"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

var mdRenderer = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
)

// RenderFragment converts a Report to an HTML fragment (no document shell).
func RenderFragment(d model.Report) string {
	src := markdown.Build(d)
	src, mathBlocks := stripDisplayMathForGoldmark(src)
	var buf bytes.Buffer
	if err := mdRenderer.Convert([]byte(src), &buf); err != nil {
		return "<pre>" + html.EscapeString(src) + "</pre>"
	}
	return injectDisplayMathHTML(buf.String(), mathBlocks)
}
