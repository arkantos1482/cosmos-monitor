package panel

import (
	"fmt"
	"html"

	"github.com/arkantos1482/cosmos-monitor/internal/feemarket"
	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func writeFeemarketSummary(w Writer, d model.Report, mode SummaryMode) {
	c := feemarket.LoadContext(d)
	transfer := feemarket.TransferCost(c.BaseFeeRaw, c.Denom)
	baseFee := d.BaseFee
	if baseFee == "" {
		baseFee = "—"
	}
	verdict := map[string]string{
		"busy": "BUSY", "quiet": "QUIET", "balanced": "BALANCED", "unknown": "UNKNOWN",
	}[c.Verdict]
	gasLine := formatGasAmount(c.GasUsed, d.ParentBlockResultsOK) + " used · " +
		formatGasAmount(c.Wanted, d.ParentBlockResultsOK) + " W"

	summaryWrapStart(w, mode, "feemarket")
	w.WriteHTML(`<div class="fee-summary">`)
	w.WriteHTML(`<div class="fee-summary__top">`)
	badgeCls := c.Badge.Class
	if badgeCls == "" {
		badgeCls = "stable"
	}
	w.WriteHTML(fmt.Sprintf(`<span class="fee-badge fee-badge--%s">%s</span>`,
		html.EscapeString(badgeCls), html.EscapeString(c.Badge.Label)))
	w.WriteHTML(fmt.Sprintf(`<span class="fee-summary__base">%s</span>`, html.EscapeString(baseFee)))
	w.WriteHTML(`</div>`)
	w.WriteHTML(`<div class="fee-summary__meta">`)
	w.WriteHTML(fmt.Sprintf(`<span class="fee-summary__pill">%s</span>`, html.EscapeString(verdict)))
	if c.UtilPct != "" {
		w.WriteHTML(fmt.Sprintf(`<span class="fee-summary__pill">%s of target</span>`, html.EscapeString(c.UtilPct)))
	}
	w.WriteHTML(fmt.Sprintf(`<span class="fee-summary__pill">transfer %s</span>`, html.EscapeString(transfer)))
	w.WriteHTML(`</div>`)
	w.WriteHTML(fmt.Sprintf(`<p class="fee-summary__gas">%s</p>`, html.EscapeString(gasLine)))
	w.WriteHTML(`</div>`)
	summaryWrapEnd(w, mode)
}

func writeFeemarket(w Writer, d model.Report) {
	w.Section("3. FEE MARKET")
	writeFeemarketSummary(w, d, SummaryEmbedded)
	w.Em("Chain-wide EIP-1559 fee market — live base fee, demand, and governance parameters.")
	writeFeemarketPage(w, d)
	w.BlankLine()
}
