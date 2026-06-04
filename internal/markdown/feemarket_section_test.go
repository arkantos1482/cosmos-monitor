package markdown

import (
	"strings"
	"testing"
)

func TestLatexUintUnlimited(t *testing.T) {
	if got := latexUint(^uint64(0)); got != `\text{unlimited}` {
		t.Fatalf("got %q", got)
	}
}

func TestWriteFeeMathMarkdown(t *testing.T) {
	var b strings.Builder
	writeFeeMathMarkdown(&b, `\[ W_{\text{stored}} = 1 \]`)
	out := b.String()
	if !strings.Contains(out, "W_{\\text{stored}}") {
		t.Fatalf("expected latex in $$ block, got %q", out)
	}
	if strings.Contains(out, `class="fee-math`) {
		t.Fatal("must not emit HTML fee-math hooks")
	}
}
