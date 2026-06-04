package terminal

import (
	"fmt"
	"io"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
	"github.com/arkantos1482/cosmos-monitor/internal/panel"
)

// Text prints the dashboard as plain text to w.
type Text struct {
	W io.Writer
}

func (t Text) Render(d model.Report) error {
	_, err := fmt.Fprint(t.W, panel.BuildText(d))
	return err
}
