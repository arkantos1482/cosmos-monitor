package terminal

import (
	"fmt"
	"io"

	"github.com/arkantos1482/cosmos-monitor/internal/markdown"
	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

// Raw prints canonical Markdown to w.
type Raw struct {
	W io.Writer
}

func (r Raw) Render(d model.Report) error {
	_, err := fmt.Fprint(r.W, markdown.Build(d))
	return err
}
