package panel

import (
	"fmt"
	"html"
	"strconv"
	"strings"

	"github.com/arkantos1482/cosmos-monitor/internal/feemarket"
	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func writeFeemarketSummary(w Writer, d model.Report, mode SummaryMode) {
	s := feemarket.LoadState(d)
	summaryWrapStart(w, mode, "feemarket")
	writeFeemarketSummaryBody(w, s, d)
	summaryWrapEnd(w, mode)
}

func writeFeemarketSummaryBody(w Writer, s feemarket.State, d model.Report) {
	baseFee := d.BaseFee
	if baseFee == "" {
		baseFee = "—"
	}
	w.WriteHTML(`<div class="fm-summary">`)
	w.WriteHTML(`<div class="fm-summary__top">`)
	w.WriteHTML(fmt.Sprintf(`<span class="fm-summary__base">%s</span>`, html.EscapeString(baseFee)))
	w.WriteHTML(fmt.Sprintf(`<span class="badge %s">%s</span>`, s.Adj.BadgeClass(), html.EscapeString(s.Adj.Label())))
	w.WriteHTML(`</div>`)
	if s.UtilPct > 0 {
		writeMiniGauge(w, "block demand vs target", s.UtilPct)
	}
	w.WriteHTML(`<div class="fm-summary__kpis">`)
	writeFmSummaryKPI(w, "mode", s.Mode)
	transfer := feemarket.TransferCost(s.BaseFeeRaw, s.Denom)
	if transfer != "—" {
		writeFmSummaryKPI(w, "21k transfer", transfer)
	}
	w.WriteHTML(`</div></div>`)
}

func writeFmSummaryKPI(w Writer, label, value string) {
	if value == "" {
		return
	}
	w.WriteHTML(fmt.Sprintf(
		`<div class="fm-summary__kpi"><span class="fm-summary__kpi-label">%s</span>`+
			`<span class="fm-summary__kpi-val">%s</span></div>`,
		html.EscapeString(label), html.EscapeString(value)))
}

func writeFeemarket(w Writer, d model.Report) {
	s := feemarket.LoadState(d)

	w.Section("3. FEE MARKET")
	writeEmbeddedSectionIntro(w, "x/feemarket module state: EIP-1559 base fee, block gas demand, on-chain parameters, and this node's local fee policy.")
	writeFeemarketSummary(w, d, SummaryEmbedded)

	w.Subsection("Live state")
	w.Hint("Parent block (H−1) execution gas and mempool limits determine W; see EIP-1559 mechanics for the formula inputs.")
	w.Row("block height", d.BlockHeight)
	w.Row("base fee", orDash(d.BaseFee))
	if d.ParentBlockResultsOK || s.GasUsed > 0 {
		w.Row("gas used", formatUint(s.GasUsed)+" gas  _(parent block execution, Σ tx gas_used — floor for W)_")
	}
	if d.ParentBlockResultsOK || s.TxGasWanted > 0 {
		w.Row("gas wanted", formatUint(s.TxGasWanted)+" gas  _(parent block mempool limits, Σ tx gas_wanted — scaled by min_gas_multiplier for W)_")
	}
	if html := fmDemandVsTargetHTML(s); html != "" {
		w.RowHTML("demand vs target", html, "W vs gas target")
	}
	if s.LastBlockFees != "" {
		w.Row("parent block fees", s.LastBlockFees+"  _(gas_used × base_fee)_")
	}
	if s.HasProjection {
		w.Row("projected next base fee", fmProjectedFee(s)+"  _(EIP-1559 projection from parent W and target)_")
	}

	w.Subsection("EIP-1559 mechanics")
	w.WriteHTML(fmMechanicsHTML(s))

	w.Subsection("Module & policy")
	w.WriteHTML(feemarketDomainCardsHTML(s, d))

	writeSectionSources(w, ViewFeemarket, d)
	w.BlankLine()
}

func fmMechanicsHTML(s feemarket.State) string {
	var b strings.Builder
	b.WriteString(`<div class="fm-mechanics">`)
	b.WriteString(`<p class="fm-mechanics__lead">Base fee tracks block <strong>demand (W) vs target</strong> using the standard EIP-1559 curve.</p>`)
	b.WriteString(`<ol class="fm-mechanics__steps">`)
	b.WriteString(`<li><strong>EndBlock</strong> — compute W = max(gas_used, gas_wanted × min_gas_multiplier) from the parent block.</li>`)
	b.WriteString(`<li><strong>BeginBlock</strong> — compare W to target (block gas limit ÷ elasticity) and adjust base fee.</li>`)
	b.WriteString(`<li><strong>CheckTx / ante</strong> — txs must meet base fee (EVM) and node minimum-gas-prices (Cosmos).</li>`)
	b.WriteString(`</ol>`)
	b.WriteString(`<pre class="fm-mechanics__formula">`)
	b.WriteString(html.EscapeString(
		"W = max(gas_used, gas_wanted × min_gas_multiplier)\n\n" +
			"if W > target:  base_fee += max(base_fee × |W−target| / target / denom, 1 unit)\n" +
			"if W < target:  base_fee -= base_fee × |W−target| / target / denom  (floor = min_gas_price)\n" +
			"if W == target: base_fee unchanged"))
	b.WriteString(`</pre>`)
	if vars := fmMechanicsVarsHTML(s); vars != "" {
		b.WriteString(`<p class="fm-mechanics__vars-title">Live values (parent block)</p>`)
		b.WriteString(vars)
	}
	if s.Mode != "" {
		b.WriteString(fmt.Sprintf(`<p class="fm-mechanics__mode">Current mode: <strong>%s</strong></p>`,
			html.EscapeString(s.Mode)))
	}
	b.WriteString(`</div>`)
	return b.String()
}

func formatUint(n uint64) string {
	return strconv.FormatUint(n, 10)
}
