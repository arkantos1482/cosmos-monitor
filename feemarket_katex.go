package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"strings"
)

func isUnlimitedBlockGas(n uint64) bool {
	return n == ^uint64(0)
}

// katexUint formats a uint64 for KaTeX (avoids int64 overflow on unlimited max_gas).
func katexUint(n uint64) string {
	if isUnlimitedBlockGas(n) {
		return `\text{unlimited}`
	}
	if n > uint64(1<<63-1) {
		return fmt.Sprintf(`%d`, n)
	}
	return katexInt(int64(n))
}

func splitLatexDisplayBlocks(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	var blocks []string
	for _, chunk := range strings.Split(s, `\]`) {
		chunk = strings.TrimSpace(chunk)
		chunk = strings.TrimPrefix(chunk, `\[`)
		chunk = strings.TrimSpace(chunk)
		if chunk != "" {
			blocks = append(blocks, chunk)
		}
	}
	return blocks
}

func writeFeeMathHTML(w io.Writer, latexParts ...string) {
	fmt.Fprint(w, `<div class="fee-math">`+"\n")
	for _, part := range latexParts {
		for _, block := range splitLatexDisplayBlocks(part) {
			writeKaTeXNode(w, block)
		}
	}
	fmt.Fprint(w, `</div>`+"\n\n")
}

func writeKaTeXNode(w io.Writer, latex string) {
	b64 := base64.StdEncoding.EncodeToString([]byte(latex))
	fmt.Fprintf(w, `<div class="fee-math-tex" data-tex-b64="%s"></div>`+"\n", b64)
}
