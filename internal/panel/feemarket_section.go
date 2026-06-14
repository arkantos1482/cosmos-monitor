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
	if s.GasWanted > 0 {
		writeFmSummaryKPI(w, "block gas (W)", formatUint(s.GasWanted))
	}
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
	w.Row("block height", d.BlockHeight)
	w.Row("base fee", orDash(d.BaseFee))
	if s.GasWanted > 0 {
		w.Row("block gas (W)", formatUint(s.GasWanted)+" gas  _(parent demand input)_")
	}
	if s.GasUsed > 0 {
		w.Row("parent gas used", formatUint(s.GasUsed)+" gas")
	}
	if html := fmDemandVsTargetHTML(s); html != "" {
		w.RowHTML("demand vs target", html, "")
	}
	if s.LastBlockFees != "" {
		w.Row("parent block fees", s.LastBlockFees+"  _(gas_used × base_fee)_")
	}
	if s.HasProjection {
		w.Row("projected next base fee", fmProjectedFee(s)+"  _(CalcGasBaseFee from module)_")
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
	b.WriteString(`<p class="fm-mechanics__lead">Base fee tracks block <strong>demand vs target</strong> — same EIP-1559 curve as Ethereum (` +
		`<code>CalcGasBaseFee</code>).</p>`)
	b.WriteString(`<ol class="fm-mechanics__steps">`)
	b.WriteString(`<li><strong>BeginBlock</strong> — read parent demand, compare to <code>target = gasLimit ÷ elasticity</code>, adjust base fee.</li>`)
	b.WriteString(`<li><strong>CheckTx / ante</strong> — txs must meet base fee (EVM) and node <code>minimum-gas-prices</code> (Cosmos).</li>`)
	b.WriteString(`</ol>`)
	b.WriteString(`<p class="fm-mechanics__cosmos"><span class="fm-mechanics__cosmos-tag">Cosmos</span> ` +
		`EndBlock stores <strong>W</strong> = max(gas_used, gas_wanted × min_gas_multiplier) so mempool pressure counts — not execution gas alone.</p>`)
	b.WriteString(`<pre class="fm-mechanics__formula">`)
	b.WriteString(html.EscapeString(
		"if W > target:  base_fee += max(base_fee × |W−target| / target / denom, 1 apmt)\n" +
			"if W < target:  base_fee -= base_fee × |W−target| / target / denom  (floor = min_gas_price)\n" +
			"if W == target: base_fee unchanged"))
	b.WriteString(`</pre>`)
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
