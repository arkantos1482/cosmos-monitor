package render

import (
	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

// Renderer turns a Report into terminal or web output.
type Renderer interface {
	Render(d model.Report) error
}
