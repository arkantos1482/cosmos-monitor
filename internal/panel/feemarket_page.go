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
		{"fee-L2", "L2 Cause"},
		{"fee-L3", "L3 Demand"},
		{"fee-L4", "L4 Timeline"},
		{"fee-L5", "L5 Precision"},
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
		buildFeeL2(c, d),
		buildFeeL3(c, d),
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

func buildFeeL2(c feemarket.Context, d model.Report) feeLevel {
	verdictLabel := map[string]string{
		"busy": "BUSY", "quiet": "QUIET", "balanced": "BALANCED", "unknown": "UNKNOWN",
	}[c.Verdict]
	verdictDetail := ""
	switch c.Verdict {
	case "busy":
		verdictDetail = "(W above target)"
	case "quiet":
		verdictDetail = "(W below target)"
	case "balanced":
		verdictDetail = "(W at target)"
	}

	maxGasVal := feemarket.FormatUint(c.BlockGasLimit)
	maxGasNote := "consensus max_gas"
	if c.UnlimitedBlockGas {
		maxGasVal = "unlimited (−1)"
	}

	nextAdj := c.NextAdj
	switch c.Badge.Label {
	case "AT FLOOR":
		nextAdj = "HOLD at floor"
	case "FALLING":
		if !c.DecreaseStep.IsNil() && c.DecreaseStep.IsPositive() {
			nextAdj = fmt.Sprintf("FALLING — Δbase ≈ %s/gas", fetch.FormatFeeDec(c.DecreaseStep, c.Denom))
		} else {
			nextAdj = "FALLING — base fee drops next block"
		}
	case "RISING":
		nextAdj = "RISING — base fee increases next block"
	case "STABLE":
		nextAdj = "STABLE — base fee holds"
	}

	rows := [][]string{
		{fmt.Sprintf("Verdict (block %s)", c.ParentBlock), verdictLabel + "  " + verdictDetail},
		{"Target", c.TargetDisplay()},
		{"max_gas (consensus)", maxGasVal + "  _(" + maxGasNote + ")_"},
		{fmt.Sprintf("Base at BeginBlock %s", c.CurrentBlock), feeAmount(d, c.BaseFee, c.BaseFeeRaw)},
		{"Next adjustment", nextAdj},
	}
	if c.UnlimitedBlockGas {
		rows = append(rows, []string{
			"Demand meter", "— (target is sentinel, not real capacity)",
			"Why quiet with sentinel", fmt.Sprintf("W (%s gas) ≪ sentinel target", feemarket.FormatUint(c.Wanted)),
		})
	} else if c.HasTarget && c.UtilPct != "" {
		rows = append(rows, []string{"Demand meter", c.UtilPct + " of target"})
	}

	var extra strings.Builder
	if !c.UnlimitedBlockGas && c.HasTarget {
		extra.WriteString(feeDemandMeter(c.LoadBarPct, c.UtilPct))
	}

	return feeLevel{
		ID:      "fee-L2",
		Title:   "L2 · Why the fee moved",
		Concept: "Fees move because last-block demand was above, below, or at the network target. Adjustment lags one block.",
		Rows:    rows,
		Extra:   extra.String(),
	}
}

func buildFeeL3(c feemarket.Context, d model.Report) feeLevel {
	mult := c.MinGasMultiplier
	if mult == "" {
		mult = "0.5"
	}
	example := fmt.Sprintf(
		`<div class="fee-example">`+
			`<div class="fee-example__title">Illustrative example: when W ≠ gas_used</div>`+
			`<p class="fee-example__note">Hypothetical numbers — not live chain data for this block.</p>`+
			`<table class="fee-table"><tbody>`+
			`<tr><th>In-block accumulator</th><td>42,000 gas</td><td class="fee-table__note">sum of tx gas limits in block</td></tr>`+
			`<tr><th>× min_gas_multiplier %s</th><td>21,000 gas</td><td></td></tr>`+
			`<tr><th>gas_used</th><td>18,000 gas</td><td class="fee-table__note">e.g. partial revert</td></tr>`+
			`<tr><th>W = max(21,000, 18,000)</th><td><strong>21,000 gas</strong></td><td class="fee-table__note">fee math uses this value</td></tr>`+
			`</tbody></table></div>`,
		html.EscapeString(mult),
	)

	return feeLevel{
		ID:      "fee-L3",
		Title:   "L3 · What the chain measured as demand",
		Concept: "Demand for the formula is W (stored block_gas), which can differ from gas_used.",
		Rows: [][]string{
			{fmt.Sprintf("Parent block %s", c.ParentBlock), "fee algorithm reads W, not raw gas_used alone"},
			{"gas_used", formatGasAmount(c.GasUsed, d.ParentBlockResultsOK)},
			{"W", formatGasAmount(c.Wanted, d.ParentBlockResultsOK)},
			{"Relationship", c.WGasUsedRelation()},
		},
		Extra: example,
	}
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
		Title:   "L4 · When each value is written",
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

	chainRows := [][]string{
		{"no_base_fee", boolStr(d.NoBaseFee)},
		{"enable_height", formatEnableHeight(c.EnableHeight)},
		{"base_fee (param store)", orDash(c.BaseFeeParam)},
		{"base_fee_change_denominator", formatParamUint(c.DenomU)},
		{"elasticity_multiplier", formatParamInt(c.Elasticity)},
		{"min_gas_price", minGasPriceL5(c)},
		{"min_gas_multiplier", orDash(c.MinGasMultiplier)},
		{"max_gas", maxGasL5(c)},
		{"max_bytes", formatParamInt(c.MaxBlockBytes)},
		{"evm_denom", orDash(c.Denom)},
		{"london_block", londonStatus(c)},
		{"min_unit_gas", "1 apmt"},
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
	extra.WriteString(feeTableHTML([]string{"Parameter", "Value"}, chainRows))
	if c.UnlimitedBlockGas && c.MaxBlockBytes > 0 {
		extra.WriteString(feeSubheadingHTML("Practical block limits (max_gas unlimited)"))
		fmt.Fprintf(&extra, `<p class="fee-level__note">max_bytes %s · block time %s · validator execution cap applies</p>`,
			feemarket.FormatUint(uint64(c.MaxBlockBytes)), orDash(c.BlockInterval))
	}
	extra.WriteString(feeSubheadingHTML("Data sources"))
	extra.WriteString(feemarketDataSourcesHint(c))
	extra.WriteString(noteCalloutHTML("Cosmos EVM uses W not gas_used; finite vs sentinel target when max_gas is −1."))

	return feeLevel{
		ID:      "fee-L5",
		Title:   "L5 · Formula, parameters, data sources",
		Concept: "Full computation, governance knobs, and provenance (node app.toml → § Infrastructure).",
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

func feemarketDataSourcesHint(c feemarket.Context) string {
	appToml := "local app.toml (APPTOML_PATH or ~/.evmd/config/app.toml)"
	return provenanceCalloutHTML(fmt.Sprintf(
		"`head height` → CometBFT GET /status; "+
			"`max_gas`, `max_bytes` → CometBFT GET /consensus_params; "+
			"`gas_used`, `W` → CometBFT GET /block_results?height=%s; "+
			"`base_fee` (BeginBlock) → CometBFT GET /block_results?height=%s; "+
			"`block interval` → CometBFT GET /block; "+
			"`base_fee` → REST GET /cosmos/evm/feemarket/v1/base_fee; "+
			"`W` (fallback) → REST GET /cosmos/evm/feemarket/v1/block_gas; "+
			"`no_base_fee`, `elasticity`, `min_gas_*`, … → REST GET /cosmos/evm/feemarket/v1/params; "+
			"`evm_denom` → REST GET /cosmos/evm/vm/v1/params; "+
			"`london_block` → REST GET /cosmos/evm/vm/v1/config; "+
			"node fee acceptance (app.toml) → %s (§ Infrastructure).",
		c.ParentBlock, c.CurrentBlock, appToml,
	))
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

func feeDemandMeter(barPct float64, label string) string {
	if label == "" {
		label = "—"
	}
	if barPct < 0 {
		barPct = 0
	}
	return fmt.Sprintf(
		`<div class="fee-meter" role="meter" aria-valuenow="%.0f" aria-valuemin="0" aria-valuemax="100" aria-label="Demand vs target">`+
			`<div class="fee-meter__label"><span>Demand vs target</span><span>%s</span></div>`+
			`<div class="fee-meter__track"><div class="fee-meter__fill" style="width:%.1f%%"></div></div>`+
			`</div>`,
		barPct, html.EscapeString(label), barPct,
	)
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
