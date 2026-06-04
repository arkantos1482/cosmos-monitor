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

func TestWriteFeeMathHTMLEscapesBackslashes(t *testing.T) {
	var b strings.Builder
	writeFeeMathHTML(&b, `\[ W_{\text{stored}} = 1 \]`)
	out := b.String()
	if strings.Contains(out, `W_{\text{stored}}`) {
		t.Fatal("raw latex must not appear in HTML")
	}
	if !strings.Contains(out, `data-tex-b64="`) {
		t.Fatal("expected base64 attribute")
	}
}
