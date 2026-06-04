package main

import (
	"fmt"
	"io"
)

func writeFeemarketSection(w io.Writer, d WebData, web bool) {
	ex := buildFeemarketExplain(d)

	fmt.Fprintf(w, "**%s**\n\n", ex.SummaryLine)

	if web {
		writeFeeMathHTML(w, ex.LatexGeneral, ex.LatexSubstituted)
	} else {
		fmt.Fprintf(w, "```text\n%s\n```\n\n", ex.TextReceipt)
	}

	writeFeemarketDiagram(w, d, web)
}
