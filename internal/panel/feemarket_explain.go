package panel

import (
	"fmt"
	"strings"

	"cosmossdk.io/math"
	"github.com/arkantos1482/cosmos-monitor/internal/fetch"
	"github.com/arkantos1482/cosmos-monitor/internal/model"
	"github.com/arkantos1482/cosmos-monitor/internal/report"
)

// FeemarketExplain holds rendered fee-market explanation (portable Markdown + KaTeX).
type FeemarketExplain struct {
	SummaryLine      string
	LatexGeneral     string
	LatexSubstituted string
	TextReceipt      string // kept for unit tests / debugging
	Verdict          string // plain English, e.g. "fee increases (↑)"
	UtilizationPct   string
	MatchOK          bool
}

func calcGasBaseFee(gasUsed, gasTarget, baseFeeChangeDenom uint64, baseFee, minUnitGas, minGasPrice math.LegacyDec) math.LegacyDec {
	if baseFee.IsNil() {
		return math.LegacyZeroDec()
	}
	if gasUsed == gasTarget {
		return baseFee
	}
	if gasTarget == 0 {
		return math.LegacyZeroDec()
	}
	num := math.LegacyNewDecFromInt(math.NewIntFromUint64(gasUsed).Sub(math.NewIntFromUint64(gasTarget)).Abs())
	num = num.Mul(baseFee)
	num = num.QuoInt(math.NewIntFromUint64(gasTarget))
	num = num.QuoInt(math.NewIntFromUint64(baseFeeChangeDenom))

	if gasUsed > gasTarget {
		baseFeeDelta := math.LegacyMaxDec(num, minUnitGas)
		return baseFee.Add(baseFeeDelta)
	}
	return math.LegacyMaxDec(baseFee.Sub(num), minGasPrice)
}

func parseLegacyDec(s string) (math.LegacyDec, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return math.LegacyDec{}, false
	}
	d, err := math.LegacyNewDecFromStr(s)
	return d, err == nil
}

func inferParentBaseFee(current math.LegacyDec, gasUsed, gasTarget, denom uint64, minUnit, minGasPrice math.LegacyDec) (math.LegacyDec, bool) {
	if gasTarget == 0 || current.IsNil() {
		return math.LegacyDec{}, false
	}
	if gasUsed == gasTarget {
		return current, true
	}
	var candidates []math.LegacyDec
	if gasUsed < gasTarget && gasTarget > 0 && denom > 0 {
		factor := math.LegacyOneDec().Sub(
			math.LegacyNewDecFromInt(math.NewIntFromUint64(gasTarget).Sub(math.NewIntFromUint64(gasUsed))).
				QuoInt(math.NewIntFromUint64(gasTarget)).
				QuoInt(math.NewIntFromUint64(denom)),
		)
		if factor.IsPositive() {
			candidates = append(candidates, current.Quo(factor))
		}
		candidates = append(candidates, current.Add(
			current.Mul(math.LegacyNewDecFromInt(math.NewIntFromUint64(gasTarget).Sub(math.NewIntFromUint64(gasUsed)))).
				QuoInt(math.NewIntFromUint64(gasTarget)).
				QuoInt(math.NewIntFromUint64(denom)),
		))
	}
	if gasUsed > gasTarget && gasTarget > 0 && denom > 0 {
		factor := math.LegacyOneDec().Add(
			math.LegacyNewDecFromInt(math.NewIntFromUint64(gasUsed).Sub(math.NewIntFromUint64(gasTarget))).
				QuoInt(math.NewIntFromUint64(gasTarget)).
				QuoInt(math.NewIntFromUint64(denom)),
		)
		if factor.IsPositive() {
			candidates = append(candidates, current.Quo(factor))
		}
		candidates = append(candidates, current.Sub(minUnit))
	}
	for _, parent := range candidates {
		if parent.IsNegative() {
			continue
		}
		got := calcGasBaseFee(gasUsed, gasTarget, denom, parent, minUnit, minGasPrice)
		if got.Equal(current) {
			return parent, true
		}
	}
	return math.LegacyDec{}, false
}

func feemarketStoredWanted(d model.Report) uint64 {
	if d.ParentBlockGasWanted > 0 {
		return d.ParentBlockGasWanted
	}
	return parseDiagramUint(d.BlockGas)
}

func feemarketGasTarget(d model.Report) (uint64, bool) {
	if d.BlockGasLimit > 0 && d.Elasticity > 0 {
		return d.BlockGasLimit / uint64(d.Elasticity), true
	}
	return 0, false
}

func isUnlimitedBlockGas(n uint64) bool {
	return n == ^uint64(0)
}

func latexUint(n uint64) string {
	if isUnlimitedBlockGas(n) {
		return `\text{unlimited}`
	}
	if n > uint64(1<<63-1) {
		return fmt.Sprintf(`%d`, n)
	}
	return latexInt(int64(n))
}

func latexTextLit(s string) string {
	r := strings.NewReplacer(
		`\`, `\textbackslash{}`,
		"{", `\{`,
		"}", `\}`,
		"#", `\#`,
		"%", `\%`,
		"&", `\&`,
	)
	return `\text{` + r.Replace(s) + `}`
}

func latexInt(n int64) string {
	return strings.ReplaceAll(report.FormatInt(n), ",", `{,}`)
}

func latexDisplayLines(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	var b strings.Builder
	for i, line := range lines {
		if i > 0 {
			b.WriteString("\n\n")
		}
		fmt.Fprintf(&b, `\[ %s \]`, line)
	}
	return b.String()
}

func feemarketVerdictPlain(wanted, target uint64, hasTarget bool) string {
	if !hasTarget {
		return "target unknown — cannot predict fee direction"
	}
	switch {
	case wanted > target:
		return "fee increases (↑) — last block was busier than target"
	case wanted < target:
		return "fee decreases (↓) — last block was quieter than target"
	default:
		return "unchanged (=) — last block matched target load"
	}
}

func feemarketVerdictLatexPlain(wanted, target uint64, hasTarget bool) string {
	if !hasTarget {
		return `\text{Cannot compare load — block gas target unknown}`
	}
	switch {
	case wanted > target:
		return `\Rightarrow \textbf{\text{Minimum fee goes up next block}} \;(\uparrow)`
	case wanted < target:
		return `\Rightarrow \textbf{\text{Minimum fee goes down next block}} \;(\downarrow)`
	default:
		return `\Rightarrow \textbf{\text{Minimum fee stays about the same}} \;(=)`
	}
}

func feemarketVerdictLatex(wanted, target uint64, hasTarget bool) string {
	if !hasTarget {
		return `\text{target unknown}`
	}
	switch {
	case wanted > target:
		return `W_{\text{stored}} > T \Rightarrow \text{fee rises }(\uparrow)`
	case wanted < target:
		return `W_{\text{stored}} < T \Rightarrow \text{fee falls }(\downarrow)`
	default:
		return `W_{\text{stored}} = T \Rightarrow \text{unchanged }(=)`
	}
}

func feemarketTechnicalFormulas() []string {
	return []string{
		`\textbf{\text{Chain formulas (x/feemarket — CalcGasBaseFee)}}`,
		`W_{\text{stored}}=\max(W_{\text{sum}}\cdot\mu,\;G_{\text{used}})\quad\text{(EndBlock, stored for next block)}`,
		`\Delta=\frac{\left|W_{\text{stored}}-T\right|\cdot B_{\text{parent}}}{T\cdot d}`,
		`B_{\text{new}}=\begin{cases}B_{\text{parent}} & W_{\text{stored}}=T\\ B_{\text{parent}}+\max(\epsilon,\,\Delta) & W_{\text{stored}}>T\\ \max(B_{\min},\,B_{\text{parent}}-\Delta) & W_{\text{stored}}<T\end{cases}`,
		`\text{Symbols: }G_{\text{used}}\text{ gas burned; }W_{\text{sum}}\text{ mempool sum; }\mu\text{ min\_gas\_multiplier; }T=\lfloor\text{max\_gas}/e\rfloor\text{; }d\text{ base\_fee\_change\_denominator; }B\text{ base fee; }\epsilon=1.`,
	}
}

func formatDecAmount(d math.LegacyDec, denom string) string {
	if d.IsNil() {
		return "—"
	}
	return fetch.FormatFeeAmount(d.String(), denom)
}

func feemarketEducationalGeneral() string {
	return latexDisplayLines([]string{
		`\textbf{\text{How PMT adjusts transaction fees (EIP-1559 style)}}`,
		`\text{Each block has a \textbf{minimum gas price} (base fee). It moves slowly—like a thermostat—based on how busy the \emph{previous} block was.}`,
		`\text{\textbf{Step 1 — Measure last block:} count gas actually used by transactions.}`,
		`\text{\textbf{Step 2 — Compare to a target:} the chain aims for blocks about half full (target = max block gas ÷ elasticity).}`,
		`\text{\textbf{Step 3 — Nudge the minimum fee:} busier than target → raise; quieter → lower (never below the chain floor).}`,
		`\text{\textbf{What you pay:}}\quad \underbrace{\text{gas used}}_{\text{computation units}} \times (\text{minimum gas price} + \text{optional priority tip})`,
	})
}

func buildFeemarketExplain(d model.Report) FeemarketExplain {
	ex := FeemarketExplain{}
	denom := diagramDenom(d)

	if d.NoBaseFee {
		ex.SummaryLine = "Fixed gas pricing — EIP-1559 auto-adjust is off (`no_base_fee`)"
		ex.TextReceipt = ex.SummaryLine
		ex.LatexGeneral = latexDisplayLines([]string{
			`\textbf{\text{Fixed minimum gas price}}`,
			`\text{The chain is \textbf{not} using EIP-1559 style auto-adjustment (\texttt{no\_base\_fee} is enabled).}`,
			`\text{Fees do not rise or fall with block congestion; wallets use the configured minimum gas price instead.}`,
		})
		ex.LatexSubstituted = ex.LatexGeneral
		return ex
	}

	wanted := feemarketStoredWanted(d)
	target, hasTarget := feemarketGasTarget(d)
	gasUsed := d.ParentBlockGasUsed

	if hasTarget && target > 0 {
		pct := float64(wanted) / float64(target) * 100
		ex.UtilizationPct = fmt.Sprintf("%.2f%%", pct)
	}

	ex.Verdict = feemarketVerdictPlain(wanted, target, hasTarget)

	ex.SummaryLine = fmt.Sprintf("Block %s — minimum gas price (base fee) %s", d.BlockHeight, d.BaseFee)
	if d.GasPrice != "" {
		ex.SummaryLine += fmt.Sprintf(" · wallet quote (eth_gasPrice) %s", fetch.FormatFeeAmount(d.GasPrice, denom))
	}
	if ex.UtilizationPct != "" {
		ex.SummaryLine += fmt.Sprintf(" · previous block load %s of target", ex.UtilizationPct)
	}

	general := append([]string{}, splitLatexDisplayBlocks(feemarketEducationalGeneral())...)
	general = append(general, feemarketTechnicalFormulas()...)
	ex.LatexGeneral = latexDisplayLines(general)

	var sub []string
	sub = append(sub, fmt.Sprintf(`\textbf{\text{Block %s — live walkthrough}}`, d.BlockHeight))
	if !d.ParentBlockResultsOK {
		sub = append(sub, `\text{⚠ Previous-block gas usage may be incomplete (CometBFT block\_results unavailable).}`)
	}

	// Step 1 — how busy was the last block?
	if gasUsed > 0 || wanted > 0 {
		load := wanted
		if load == 0 {
			load = gasUsed
		}
		sub = append(sub, fmt.Sprintf(
			`\text{\textbf{Step 1 — Previous block workload:}}\quad \text{gas used}=%s`,
			latexInt(int64(gasUsed)),
		))
		if d.MinGasMultiplier != "" && d.MinGasMultiplier != "1" {
			sub = append(sub, fmt.Sprintf(
				`\text{Recorded demand for fee math}=\max(\text{mempool}\times %s,\,\text{gas used})=%s`,
				d.MinGasMultiplier, latexInt(int64(wanted)),
			))
		} else if wanted > 0 && wanted != gasUsed {
			sub = append(sub, fmt.Sprintf(
				`\text{Recorded demand for fee math}=%s`,
				latexInt(int64(wanted)),
			))
		}
	} else {
		sub = append(sub, `\text{\textbf{Step 1 — Previous block workload:}}\quad \text{no gas usage data yet}`)
	}

	// Step 2 — target & utilization
	if hasTarget {
		sub = append(sub, fmt.Sprintf(
			`\text{\textbf{Step 2 — Target “half-full” capacity:}}\quad %s \text{ gas}\quad(\text{max block gas }%s \div \text{elasticity }%s)`,
			latexUint(target), latexUint(d.BlockGasLimit), latexInt(d.Elasticity),
		))
		if ex.UtilizationPct != "" {
			sub = append(sub, fmt.Sprintf(
				`\text{Load vs target}=\frac{%s}{%s}\approx %s`,
				latexInt(int64(wanted)), latexUint(target), latexTextLit(ex.UtilizationPct),
			))
		}
		sub = append(sub, feemarketVerdictLatexPlain(wanted, target, true))
		sub = append(sub, feemarketVerdictLatex(wanted, target, true))
	} else {
		sub = append(sub, `\text{\textbf{Step 2 — Target capacity:}}\quad \text{unknown (need max block gas from consensus)}`)
	}

	sub = append(sub, `\text{\textbf{Step 3 — Minimum gas price for this block:}}`)
	if d.BaseFee != "" {
		sub = append(sub, fmt.Sprintf(`\text{Base fee (chain minimum)}=%s`, latexTextLit(d.BaseFee)))
	}
	if d.GasPrice != "" {
		sub = append(sub, fmt.Sprintf(
			`\text{eth\_gasPrice (typical wallet quote)}=%s`,
			latexTextLit(fetch.FormatFeeAmount(d.GasPrice, denom)),
		))
	}

	sub = append(sub, `\textbf{\text{Live substitution into chain formulas}}`)
	if !d.ParentBlockResultsOK {
		sub = append(sub, `\text{⚠ parent block\_results unavailable — } G_{\text{used}} \text{ may be incomplete}`)
	}

	var params []string
	if hasTarget && d.BlockGasLimit > 0 && d.Elasticity > 0 {
		params = append(params, fmt.Sprintf(
			`T = \left\lfloor\frac{%s}{%s}\right\rfloor = %s`,
			latexUint(d.BlockGasLimit), latexInt(d.Elasticity), latexUint(target),
		))
	}
	if d.BaseFeeChangeDenominator > 0 {
		params = append(params, fmt.Sprintf(`d = %s`, latexInt(d.BaseFeeChangeDenominator)))
	}
	if d.MinGasMultiplier != "" {
		params = append(params, fmt.Sprintf(`\mu = %s`, d.MinGasMultiplier))
	}
	params = append(params, `\epsilon = 1`)
	if d.MinGasPrice != "" {
		params = append(params, fmt.Sprintf(`B_{\min} = %s`, latexTextLit(d.MinGasPrice)))
	}
	if len(params) > 0 {
		sub = append(sub, strings.Join(params, `,\quad `))
	} else if !hasTarget && d.Elasticity > 0 {
		sub = append(sub, `\text{⚠ block gas limit unknown — } T \text{ not shown (need max\_gas from consensus)}`)
	}

	mu := d.MinGasMultiplier
	if mu == "" {
		mu = `1`
	}
	sub = append(sub, fmt.Sprintf(
		`G_{\text{used}} = %s,\quad W_{\text{stored}} = \max(W_{\text{sum}} \cdot %s,\, G_{\text{used}}) = %s`,
		latexInt(int64(gasUsed)), mu, latexInt(int64(wanted)),
	))

	current, okCurrent := parseLegacyDec(d.BaseFeeRaw)
	if okCurrent && hasTarget && d.BaseFeeChangeDenominator > 0 {
		minGasPrice := math.LegacyZeroDec()
		if mp, ok := parseLegacyDec(d.MinGasPriceRaw); ok {
			minGasPrice = mp
		}
		minUnit := math.LegacyOneDec()
		denomU := uint64(d.BaseFeeChangeDenominator)
		parent, okParent := inferParentBaseFee(current, wanted, target, denomU, minUnit, minGasPrice)
		if okParent {
			delta := math.LegacyNewDecFromInt(math.NewIntFromUint64(wanted).Sub(math.NewIntFromUint64(target)).Abs())
			delta = delta.Mul(parent).QuoInt(math.NewIntFromUint64(target)).QuoInt(math.NewIntFromUint64(denomU))
			parentLit := latexTextLit(formatDecAmount(parent, denom))
			deltaLit := latexTextLit(formatDecAmount(delta, denom))
			currentLit := latexTextLit(formatDecAmount(current, denom))
			sub = append(sub, fmt.Sprintf(
				`\Delta = \frac{\left|%s - %s\right| \cdot %s}{%s \cdot %s} = %s`,
				latexUint(wanted), latexUint(target), parentLit,
				latexUint(target), latexUint(denomU), deltaLit,
			))
			sub = append(sub, fmt.Sprintf(`B_{\text{parent}} = %s`, parentLit))
			switch {
			case wanted == target:
				sub = append(sub, fmt.Sprintf(`B_{\text{new}} = B_{\text{parent}} = %s`, currentLit))
			case wanted > target:
				sub = append(sub, fmt.Sprintf(
					`B_{\text{new}} = B_{\text{parent}} + \max(\epsilon,\, \Delta) = %s + \max(1,\, %s) = %s`,
					parentLit, deltaLit, currentLit,
				))
			default:
				bMin := latexTextLit(formatDecAmount(minGasPrice, denom))
				sub = append(sub, fmt.Sprintf(
					`B_{\text{new}} = \max(B_{\min},\, B_{\text{parent}} - \Delta) = \max(%s,\, %s - %s) = %s`,
					bMin, parentLit, deltaLit, currentLit,
				))
			}
			ex.MatchOK = calcGasBaseFee(wanted, target, denomU, parent, minUnit, minGasPrice).Equal(current)
			if ex.MatchOK {
				sub = append(sub, `\checkmark\;\text{recomputed base fee matches chain}`)
			} else {
				sub = append(sub, `\text{⚠ recomputed base fee mismatch vs chain}`)
			}
		} else {
			sub = append(sub, fmt.Sprintf(
				`B_{\text{new}} = %s \quad\text{(could not infer } B_{\text{parent}}\text{ from current fee)}`,
				latexTextLit(d.BaseFee),
			))
		}
	} else if d.BaseFee != "" {
		sub = append(sub, fmt.Sprintf(`B_{\text{new}} = %s`, latexTextLit(d.BaseFee)))
	}

	ex.LatexSubstituted = latexDisplayLines(sub)

	var lines []string
	lines = append(lines, ex.SummaryLine)
	lines = append(lines, "")
	if !d.ParentBlockResultsOK {
		lines = append(lines, "⚠ parent block_results unavailable — using REST block_gas only")
	}
	if d.ParentBlockGasUsed > 0 {
		lines = append(lines, fmt.Sprintf("Parent block gas used: %s", report.FormatInt(int64(d.ParentBlockGasUsed))))
	}
	if hasTarget {
		lines = append(lines, fmt.Sprintf("Recorded demand: %s   Target: %s   (%s)",
			report.FormatInt(int64(wanted)), report.FormatInt(int64(target)), ex.Verdict))
	}
	if d.MinGasMultiplier != "" {
		lines = append(lines, fmt.Sprintf("Recorded demand = max(mempool sum × %s, gas_used)", d.MinGasMultiplier))
	}
	lines = append(lines, "")
	lines = append(lines, "CalcGasBaseFee (BeginBlock):")
	lines = append(lines, "  Δ = |W_stored − T| × B_parent / T / d")
	if d.BaseFee != "" {
		lines = append(lines, "  B_new = "+d.BaseFee)
	}
	if ex.MatchOK {
		lines = append(lines, "  ✓ matches chain base_fee")
	}
	ex.TextReceipt = strings.Join(lines, "\n")

	return ex
}
