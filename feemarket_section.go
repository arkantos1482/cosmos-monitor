package main

import (
	"fmt"
	"io"
)

func writeFeemarketSection(w io.Writer, d WebData, web bool) {
	ex := buildFeemarketExplain(d)

	fmt.Fprintf(w, "**%s**\n\n", ex.SummaryLine)

	if web {
		fmt.Fprintf(w, `<div class="fee-math">`+"\n")
		fmt.Fprintf(w, "%s\n\n", ex.LatexGeneral)
		fmt.Fprintf(w, "%s\n", ex.LatexSubstituted)
		fmt.Fprintf(w, `</div>`+"\n\n")
	} else {
		fmt.Fprintf(w, "```text\n%s\n```\n\n", ex.TextReceipt)
	}

	writeFeemarketDiagram(w, d, web)
}
