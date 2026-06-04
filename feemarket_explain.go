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

func katexEscape(s string) string {
	s = strings.ReplaceAll(s, "\\", `\textbackslash{}`)
	s = strings.ReplaceAll(s, "_", `\_`)
	return s
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
	if !d.ParentBlockResultsOK {
		sub = append(sub, `\text{⚠ parent block\_results unavailable — using REST block\_gas only}`)
	}
	if d.MinGasMultiplier != "" && gasUsed > 0 {
		sub = append(sub, fmt.Sprintf(
			`W_{\text{stored}} = \max(W_{\text{sum}} \cdot %s,\; %s) = %s`,
			katexEscape(d.MinGasMultiplier),
			fmtInt(int64(gasUsed)),
			fmtInt(int64(wanted)),
		))
	} else if d.MinGasMultiplier != "" {
		sub = append(sub, fmt.Sprintf(`W_{\text{stored}} = \max(W_{\text{sum}} \cdot %s,\; G_{\text{used}})`, katexEscape(d.MinGasMultiplier)))
		sub = append(sub, fmt.Sprintf(`W_{\text{stored}} = %s \text{ (from chain store)}`, fmtInt(int64(wanted))))
	} else {
		sub = append(sub, fmt.Sprintf(`W_{\text{stored}} = %s`, fmtInt(int64(wanted))))
	}
	if hasTarget {
		sub = append(sub, fmt.Sprintf(
			`T = %s \div %d = %s \;\Rightarrow\; %s`,
			fmtInt(int64(d.BlockGasLimit)), d.Elasticity, fmtInt(int64(target)), katexEscape(ex.Verdict),
		))
	}

	current, okCurrent := parseLegacyDec(d.BaseFeeRaw)
	if okCurrent && hasTarget && d.BaseFeeChangeDenominator > 0 {
		minGasPrice := math.LegacyZeroDec()
		if mp, ok := parseLegacyDec(d.MinGasPriceRaw); ok {
			minGasPrice = mp
		}
		minUnit := math.LegacyOneDec()
		parent, okParent := inferParentBaseFee(current, wanted, target, uint64(d.BaseFeeChangeDenominator), minUnit, minGasPrice)
		if okParent {
			delta := math.LegacyNewDecFromInt(math.NewIntFromUint64(wanted).Sub(math.NewIntFromUint64(target)).Abs())
			delta = delta.Mul(parent).QuoInt(math.NewIntFromUint64(target)).QuoInt(math.NewIntFromUint64(uint64(d.BaseFeeChangeDenominator)))
			sub = append(sub, fmt.Sprintf(`B_{\text{parent}} = %s`, katexEscape(formatDecAmount(parent, denom))))
			sub = append(sub, fmt.Sprintf(`\Delta = %s`, katexEscape(formatDecAmount(delta, denom))))
			sub = append(sub, fmt.Sprintf(`B_{\text{new}} = %s`, katexEscape(formatDecAmount(current, denom))))
			ex.MatchOK = calcGasBaseFee(wanted, target, uint64(d.BaseFeeChangeDenominator), parent, minUnit, minGasPrice).Equal(current)
			if ex.MatchOK {
				sub = append(sub, `\checkmark \text{ matches chain base\_fee}`)
			} else {
				sub = append(sub, `\text{⚠ mismatch vs chain base\_fee}`)
			}
		} else {
			sub = append(sub, fmt.Sprintf(`B_{\text{new}} = %s \text{ (could not infer } B_{\text{parent}} \text{)}`, katexEscape(d.BaseFee)))
		}
	} else if d.BaseFee != "" {
		sub = append(sub, fmt.Sprintf(`B_{\text{new}} = %s`, katexEscape(d.BaseFee)))
	}

	ex.LatexSubstituted = `\begin{aligned}` + strings.Join(sub, ` \\ `) + `\end{aligned}`

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
