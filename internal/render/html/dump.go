package html

import (
	"fmt"
	"io"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
	"github.com/arkantos1482/cosmos-monitor/internal/panel"
)

// Dump writes an HTML fragment (no document shell) for --dump.
type Dump struct {
	W    io.Writer
	Opts panel.Options
}

func (d Dump) Render(rep model.Report) error {
	_, err := fmt.Fprint(d.W, RenderFragmentWithOptions(rep, d.Opts))
	return err
}
