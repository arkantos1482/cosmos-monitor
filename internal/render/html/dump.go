package html

import (
	"fmt"
	"io"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

// Dump writes an HTML fragment (no document shell) for --dump.
type Dump struct {
	W io.Writer
}

func (d Dump) Render(rep model.Report) error {
	_, err := fmt.Fprint(d.W, RenderFragment(rep))
	return err
}
