package panel

import (
	"fmt"
	"html"
	"strings"

	"github.com/arkantos1482/cosmos-monitor/internal/feemarket"
	"github.com/arkantos1482/cosmos-monitor/internal/fetch"
	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func writeFeemarketPage(w Writer, d model.Report) {
	c := feemarket.LoadContext(d)
	w.Subsection("Chain state & parameters")
	w.WriteHTML(feemarketDomainCardsHTML(d))
	levels := buildFeeLevels(c, d)
	writeFeeNav(w, c)
	for _, lv := range levels {
		writeFeeLevel(w, lv)
	}
}

func writeFeeNav(w Writer, c feemarket.Context) {
	meta := fmt.Sprintf("Block %s · refresh 5s", c.CurrentBlock)
	var b strings.Builder
	b.WriteString(`<nav class="fee-nav" aria-label="Fee market levels">`)
	for _, link := range []struct{ id, label string }{
		{"fee-L1", "L1 Outcome"},
		{"fee-L4", "L2 Timeline"},
		{"fee-L5", "L3 Formula"},
	} {
		fmt.Fprintf(&b, `<a class="fee-nav__link" href="#%s">%s</a>`, link.id, html.EscapeString(link.label))
	}
	fmt.Fprintf(&b, `<span class="fee-nav__meta">%s</span>`, html.EscapeString(meta))
	b.WriteString(`</nav>`)
	w.WriteHTML(b.String())
}

func buildFeeLevels(c feemarket.Context, d model.Report) []feeLevel {
	return []feeLevel{
		buildFeeL1(c, d),
		buildFeeL4(c, d),
		buildFeeL5(c, d),
	}
}

func buildFeeL1(c feemarket.Context, d model.Report) feeLevel {
	transfer := feemarket.TransferCost(c.BaseFeeRaw, c.Denom)

	lv := feeLevel{
		ID:      "fee-L1",
		Title:   "L1 · What you pay now",
		Concept: "Current minimum gas price and practical transfer cost.",
		Badge:   c.Badge,
		Footnote: c.L1Footnote,
		Rows: [][]string{
			{fmt.Sprintf("Base fee (block %s)", c.CurrentBlock), feeAmount(d, c.BaseFee, c.BaseFeeRaw)},
			{"Simple transfer (21,000 gas)", transfer},
		},
	}
	if c.FeesDisabled {
		lv.Banner = fmt.Sprintf("Fees disabled below block %s (enable_height).", formatInt(c.EnableHeight))
	}
	return lv
}

func buildFeeL4(c feemarket.Context, d model.Report) feeLevel {
	mult := c.MinGasMultiplier
	if mult == "" {
		mult = "1"
	}
	steps := []struct{ label, body string }{
		{
			fmt.Sprintf("During block %s", c.ParentBlock),
			"in-block gas accumulator += tx gas limits at ante (in-block only)",
		},
		{
			fmt.Sprintf("EndBlock %s", c.ParentBlock),
			fmt.Sprintf("W = max(acc×%s, gas_used) = %s stored", mult, feemarket.FormatUint(c.Wanted)),
		},
		{
			fmt.Sprintf("BeginBlock %s", c.CurrentBlock),
			fmt.Sprintf("base_fee := f(W, target) → %s", feeAmount(d, c.BaseFee, c.BaseFeeRaw)),
		},
		{
			fmt.Sprintf("During block %s", c.CurrentBlock),
			"txs pay ≥ base_fee; CometBFT mempool NOT in formula",
		},
	}
	var timeline strings.Builder
	timeline.WriteString(`<ol class="fee-timeline">`)
	for _, step := range steps {
		fmt.Fprintf(&timeline, `<li class="fee-timeline__step"><span class="fee-timeline__when">%s</span><span class="fee-timeline__what">%s</span></li>`,
			html.EscapeString(step.label), inlineHTML(step.body))
	}
	timeline.WriteString(`</ol>`)

	pools := `<div class="fee-pools"><div class="fee-pools__title">Three pools</div>` +
		`<table class="fee-table fee-table--pools"><thead><tr>` +
		`<th>CometBFT mempool</th><th>In-block accumulator</th><th>W (block_gas store)</th>` +
		`</tr></thead><tbody><tr>` +
		`<td>pending txs<br><em>NOT in fee math</em></td>` +
		`<td>sum gas limits in block<br><em>feeds W via multiplier</em></td>` +
		`<td>EndBlock persistence<br><em>read next BeginBlock</em></td>` +
		`</tr></tbody></table></div>`

	return feeLevel{
		ID:      "fee-L4",
		Title:   "L2 · When each value is written",
		Concept: "Block lifecycle timing: EndBlock stores W; BeginBlock applies base fee.",
		Extra:   timeline.String() + pools,
	}
}

func buildFeeL5(c feemarket.Context, d model.Report) feeLevel {
	var formula, deltaNote string
	if !c.NoBaseFee && c.HasTarget && c.DenomU > 0 {
		base := c.BaseFeeRaw
		if base == "" {
			base = c.BaseFee
		}
		wStr := feemarket.FormatUint(c.Wanted)
		tStr := c.TargetDisplay()
		formula = fmt.Sprintf(
			"|Δbase| ≤ base × |W − target| / (target × denom)\n"+
				"        ≤ %s × |%s − %s| / (%s × %d)",
			base, wStr, tStr, tStr, c.DenomU,
		)
		if !c.DecreaseStep.IsNil() && c.DecreaseStep.IsPositive() {
			deltaNote = fmt.Sprintf("→ computed decrease step %s · Badge: %s", fetch.FormatFeeDec(c.DecreaseStep, c.Denom), c.Badge.Label)
		}
	}

	var extra strings.Builder
	if formula != "" {
		extra.WriteString(`<pre class="fee-formula"><code>` + html.EscapeString(formula) + `</code></pre>`)
		if deltaNote != "" {
			fmt.Fprintf(&extra, `<p class="fee-level__note">%s</p>`, inlineHTML(deltaNote))
		}
	}
	if c.Verify != "" {
		fmt.Fprintf(&extra, `<p class="fee-level__note">Verify: %s</p>`, html.EscapeString(c.Verify))
	}
	extra.WriteString(feeSubheadingHTML("Chain parameters (governance / consensus)"))
	extra.WriteString(feeChainParamsDomainHTML(c, d))
	if c.UnlimitedBlockGas && c.MaxBlockBytes > 0 {
		extra.WriteString(feeSubheadingHTML("Practical block limits (max_gas unlimited)"))
		fmt.Fprintf(&extra, `<p class="fee-level__note">max_bytes %s · block time %s · validator execution cap applies</p>`,
			feemarket.FormatUint(uint64(c.MaxBlockBytes)), orDash(c.BlockInterval))
	}
	extra.WriteString(noteCalloutHTML("Cosmos EVM uses W not gas_used; finite vs sentinel target when max_gas is −1."))

	return feeLevel{
		ID:      "fee-L5",
		Title:   "L3 · Formula & parameters",
		Concept: "Full computation and governance knobs for the chain-wide fee market.",
		Extra:   extra.String(),
	}
}

func maxGasL5(c feemarket.Context) string {
	if c.UnlimitedBlockGas {
		return "unlimited (−1 → MaxUint64)"
	}
	if c.BlockGasLimit == 0 {
		return "—"
	}
	return feemarket.FormatUint(c.BlockGasLimit)
}

func formatGasAmount(n uint64, resultsOK bool) string {
	if !resultsOK && n == 0 {
		return "—"
	}
	return feemarket.FormatUint(n) + " gas"
}

func formatEnableHeight(n int64) string {
	if n == 0 {
		return "0 (genesis)"
	}
	return feemarket.FormatUint(uint64(n))
}

func formatParamInt(n int64) string {
	if n == 0 {
		return "—"
	}
	return feemarket.FormatUint(uint64(n))
}

func formatParamUint(n uint64) string {
	if n == 0 {
		return "—"
	}
	return feemarket.FormatUint(n)
}

func minGasPriceL5(c feemarket.Context) string {
	if c.MinGasPrice != "" {
		return c.MinGasPrice
	}
	if c.MinGasPriceRaw != "" {
		return fetch.FormatFeeAmount(c.MinGasPriceRaw, c.Denom)
	}
	return "—"
}

func londonStatus(c feemarket.Context) string {
	if c.HardforkLondon == "" {
		return "—"
	}
	if c.HardforkLondon == "0" {
		return c.HardforkLondon + " (active at genesis)"
	}
	return c.HardforkLondon
}

func feeAmount(d model.Report, display, raw string) string {
	if display == "" || display == "—" {
		return "—"
	}
	denom := feemarket.LoadContext(d).Denom
	if raw != "" {
		return fetch.FormatFeeAmount(raw, denom)
	}
	return display
}

func formatInt(n int64) string {
	if n == 0 {
		return "0"
	}
	return feemarket.FormatUint(uint64(n))
}

func boolStr(v bool) string {
	if v {
		return "true"
	}
	return "false"
}

func orDash(s string) string {
	if s == "" {
		return "—"
	}
	return s
}
