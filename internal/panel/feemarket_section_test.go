package panel

import (
	"strings"
	"testing"
)

func TestLatexUintUnlimited(t *testing.T) {
	if got := latexUint(^uint64(0)); got != `\text{unlimited}` {
		t.Fatalf("got %q", got)
	}
}

func TestWriteFeeMathHTML(t *testing.T) {
	var b strings.Builder
	w := newWriter(&b)
	w.MathLatex(`\[ W_{\text{stored}} = 1 \]`)
	out := b.String()
	if !strings.Contains(out, `math-line`) || !strings.Contains(out, `math-panel`) {
		t.Fatalf("expected math-panel/math-line, got %q", out)
	}
}
