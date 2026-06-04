package main

import (
	"fmt"
	"strings"

	"cosmossdk.io/math"
	"github.com/arkantos1482/cosmos-monitor/fetch"
)

// FeemarketExplain holds rendered fee-market explanation for web and terminal.
type FeemarketExplain struct {
	SummaryLine      string
	LatexGeneral     string
	LatexSubstituted string
	TextReceipt      string
	Verdict          string // "↑" "↓" "="
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
	// Closed-form candidates, verified by forward CalcGasBaseFee.
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

func feemarketStoredWanted(d WebData) uint64 {
	if d.ParentBlockGasWanted > 0 {
		return d.ParentBlockGasWanted
	}
	return parseDiagramUint(d.BlockGas)
}

func feemarketGasTarget(d WebData) (uint64, bool) {
	if d.BlockGasLimit > 0 && d.Elasticity > 0 {
		return d.BlockGasLimit / uint64(d.Elasticity), true
	}
	return 0, false
}

func katexTextLit(s string) string {
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

func katexInt(n int64) string {
	return strings.ReplaceAll(fmtInt(n), ",", `{,}`)
}

func katexDisplayAligned(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	return `\[ \begin{aligned}` + strings.Join(lines, ` \\ `) + ` \end{aligned} \]`
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

func formatDecAmount(d math.LegacyDec, denom string) string {
	if d.IsNil() {
		return "—"
	}
	return fetch.FormatFeeAmount(d.String(), denom)
}

func buildFeemarketExplain(d WebData) FeemarketExplain {
	ex := FeemarketExplain{}
	denom := diagramDenom(d)

	if d.NoBaseFee {
		ex.SummaryLine = "EIP-1559 disabled (`no_base_fee`)"
		ex.TextReceipt = ex.SummaryLine
		ex.LatexGeneral = `\text{EIP-1559 disabled (no\_base\_fee)}`
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

	switch {
	case !hasTarget:
		ex.Verdict = "?"
	case wanted > target:
		ex.Verdict = "fee increases (↑)"
	case wanted < target:
		ex.Verdict = "fee decreases (↓)"
	default:
		ex.Verdict = "unchanged (=)"
	}

	ex.SummaryLine = fmt.Sprintf("Block %s — base fee %s", d.BlockHeight, d.BaseFee)
	if d.GasPrice != "" {
		ex.SummaryLine += fmt.Sprintf(" · eth_gasPrice %s", fetch.FormatFeeAmount(d.GasPrice, denom))
	}
	if ex.UtilizationPct != "" {
		ex.SummaryLine += fmt.Sprintf(" · parent block %s full", ex.UtilizationPct)
	}

	ex.LatexGeneral = strings.TrimSpace(`
\[
W_{\text{stored}} = \max\left(W_{\text{sum}} \cdot \mu,\; G_{\text{used}}\right)
\]
\[
\Delta = \frac{\left|W_{\text{stored}} - T\right| \cdot B_{\text{parent}}}{T \cdot d}
\]
\[
B_{\text{new}} = \begin{cases}
B_{\text{parent}} & W_{\text{stored}} = T \\
B_{\text{parent}} + \max(\epsilon,\, \Delta) & W_{\text{stored}} > T \\
\max(B_{\min},\; B_{\text{parent}} - \Delta) & W_{\text{stored}} < T
\end{cases}
\]
\[
T = \left\lfloor \frac{\text{max\_gas}}{e} \right\rfloor,\quad
d = \text{change denom},\quad \mu = \text{min\_gas\_multiplier},\quad
\epsilon = 1\ \text{unit},\quad B_{\min} = \text{min\_gas\_price}
\]`)

	var sub []string
	sub = append(sub, fmt.Sprintf(`&\textbf{\text{Block %s — live substitution}}`, d.BlockHeight))
	if !d.ParentBlockResultsOK {
		sub = append(sub, `&\text{⚠ parent block\_results unavailable — } G_{\text{used}} \text{ may be incomplete}`)
	}

	// Chain parameters (replaces symbolic legend with live values).
	var params []string
	if hasTarget && d.BlockGasLimit > 0 && d.Elasticity > 0 {
		params = append(params, fmt.Sprintf(
			`T = \left\lfloor\frac{%s}{%s}\right\rfloor = %s`,
			katexInt(int64(d.BlockGasLimit)), katexInt(d.Elasticity), katexInt(int64(target)),
		))
	}
	if d.BaseFeeChangeDenominator > 0 {
		params = append(params, fmt.Sprintf(`d = %s`, katexInt(d.BaseFeeChangeDenominator)))
	}
	if d.MinGasMultiplier != "" {
		params = append(params, fmt.Sprintf(`\mu = %s`, d.MinGasMultiplier))
	}
	params = append(params, `\epsilon = 1`)
	if d.MinGasPrice != "" {
		params = append(params, fmt.Sprintf(`B_{\min} = %s`, katexTextLit(d.MinGasPrice)))
	}
	if len(params) > 0 {
		sub = append(sub, "&"+strings.Join(params, `,\quad `))
	}

	// EndBlock: W_stored = max(W_sum · μ, G_used).
	mu := d.MinGasMultiplier
	if mu == "" {
		mu = `1`
	}
	sub = append(sub, fmt.Sprintf(
		`&G_{\text{used}} = %s,\quad W_{\text{stored}} = \max(W_{\text{sum}} \cdot %s,\; G_{\text{used}}) = %s`,
		katexInt(int64(gasUsed)), mu, katexInt(int64(wanted)),
	))

	if hasTarget {
		sub = append(sub, "&"+feemarketVerdictLatex(wanted, target, true))
	}

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
			parentLit := katexTextLit(formatDecAmount(parent, denom))
			deltaLit := katexTextLit(formatDecAmount(delta, denom))
			currentLit := katexTextLit(formatDecAmount(current, denom))
			sub = append(sub, fmt.Sprintf(
				`&\Delta = \frac{\left|%s - %s\right| \cdot %s}{%s \cdot %s} = %s`,
				katexInt(int64(wanted)), katexInt(int64(target)), parentLit,
				katexInt(int64(target)), katexInt(int64(denomU)), deltaLit,
			))
			sub = append(sub, fmt.Sprintf(`&B_{\text{parent}} = %s`, parentLit))
			switch {
			case wanted == target:
				sub = append(sub, fmt.Sprintf(`&B_{\text{new}} = B_{\text{parent}} = %s`, currentLit))
			case wanted > target:
				sub = append(sub, fmt.Sprintf(
					`&B_{\text{new}} = B_{\text{parent}} + \max(\epsilon,\, \Delta) = %s + \max(1,\, %s) = %s`,
					parentLit, deltaLit, currentLit,
				))
			default:
				bMin := katexTextLit(formatDecAmount(minGasPrice, denom))
				sub = append(sub, fmt.Sprintf(
					`&B_{\text{new}} = \max(B_{\min},\, B_{\text{parent}} - \Delta) = \max(%s,\, %s - %s) = %s`,
					bMin, parentLit, deltaLit, currentLit,
				))
			}
			ex.MatchOK = calcGasBaseFee(wanted, target, denomU, parent, minUnit, minGasPrice).Equal(current)
			if ex.MatchOK {
				sub = append(sub, `&\checkmark\;\text{recomputed base fee matches chain}`)
			} else {
				sub = append(sub, `&\text{⚠ recomputed base fee mismatch vs chain}`)
			}
		} else {
			sub = append(sub, fmt.Sprintf(
				`&B_{\text{new}} = %s \quad\text{(could not infer } B_{\text{parent}}\text{ from current fee)}`,
				katexTextLit(d.BaseFee),
			))
		}
	} else if d.BaseFee != "" {
		sub = append(sub, fmt.Sprintf(`&B_{\text{new}} = %s`, katexTextLit(d.BaseFee)))
	}

	ex.LatexSubstituted = katexDisplayAligned(sub)

	// Plain-text receipt (terminal).
	var lines []string
	lines = append(lines, ex.SummaryLine)
	lines = append(lines, "")
	if !d.ParentBlockResultsOK {
		lines = append(lines, "⚠ parent block_results unavailable — using REST block_gas only")
	}
	if d.ParentBlockGasUsed > 0 {
		lines = append(lines, fmt.Sprintf("Parent block gas used (sum txs): %s", fmtInt(int64(d.ParentBlockGasUsed))))
	}
	if hasTarget {
		lines = append(lines, fmt.Sprintf("W_stored = %s   T = %s   (%s)", fmtInt(int64(wanted)), fmtInt(int64(target)), ex.Verdict))
	}
	if d.MinGasMultiplier != "" {
		lines = append(lines, fmt.Sprintf("EndBlock: W_stored = max(W_sum × %s, gas_used)", d.MinGasMultiplier))
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
