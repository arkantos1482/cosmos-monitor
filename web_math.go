package main

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"
)

// displayMathBlockRE matches portable $$ display blocks (blank line delimited).
var displayMathBlockRE = regexp.MustCompile(`(?s)\n\$\$\n(.*?)\n\$\$\n`)

const mathPlaceholderPrefix = "PMTOP_MATH_BLOCK_"

// stripDisplayMathForGoldmark replaces $$ blocks with placeholders so goldmark
// does not split LaTeX across <p> tags. Terminal markdown is unchanged.
func stripDisplayMathForGoldmark(md string) (string, []string) {
	var blocks []string
	out := displayMathBlockRE.ReplaceAllStringFunc(md, func(match string) string {
		sub := displayMathBlockRE.FindStringSubmatch(match)
		if len(sub) < 2 {
			return match
		}
		idx := len(blocks)
		blocks = append(blocks, sub[1])
		return fmt.Sprintf("\n\n%s%d\n\n", mathPlaceholderPrefix, idx)
	})
	return out, blocks
}

func injectDisplayMathHTML(fragmentHTML string, blocks []string) string {
	for i, tex := range blocks {
		placeholder := mathPlaceholderPrefix + fmt.Sprintf("%d", i)
		b64 := base64.StdEncoding.EncodeToString([]byte(tex))
		div := fmt.Sprintf(`<div class="math-display" data-tex-b64="%s"></div>`, b64)
		fragmentHTML = strings.ReplaceAll(fragmentHTML, "<p>"+placeholder+"</p>", div)
		fragmentHTML = strings.Replace(fragmentHTML, placeholder, div, 1)
	}
	return fragmentHTML
}
