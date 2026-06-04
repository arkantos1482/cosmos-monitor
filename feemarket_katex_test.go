package main

import (
	"strings"
	"testing"
)

func TestKatexUintUnlimited(t *testing.T) {
	if got := katexUint(^uint64(0)); got != `\text{unlimited}` {
		t.Fatalf("got %q", got)
	}
}

func TestSplitLatexDisplayBlocks(t *testing.T) {
	blocks := splitLatexDisplayBlocks(`\[ a \]

\[ b \]`)
	if len(blocks) != 2 || blocks[0] != "a" || blocks[1] != "b" {
		t.Fatalf("got %#v", blocks)
	}
}

func TestWriteFeeMathMarkdown(t *testing.T) {
	var b strings.Builder
	writeFeeMathMarkdown(&b, `\[ W_{\text{stored}} = 1 \]`)
	out := b.String()
	if !strings.Contains(out, "$$\nW_{\\text{stored}} = 1\n$$") && !strings.Contains(out, "W_{\\text{stored}}") {
		t.Fatalf("expected latex in $$ block, got %q", out)
	}
	if strings.Contains(out, `class="fee-math`) {
		t.Fatal("must not emit HTML fee-math hooks")
	}
}
