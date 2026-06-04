package panel

import (
	"fmt"
	"strings"

	"cosmossdk.io/math"
	"github.com/arkantos1482/cosmos-monitor/internal/fetch"
	"github.com/arkantos1482/cosmos-monitor/internal/model"
	"github.com/arkantos1482/cosmos-monitor/internal/report"
)

// FeemarketExplain holds structured fee-market dashboard data (no LaTeX).
type FeemarketExplain struct {
	SummaryLine       string
	Hint              string
	Verdict           string
	NextAdj           string // ↑, ↓, or =
	TrafficLabel      string // FEE RISING / FEE FALLING / STABLE
	TrafficClass      string // rising | falling | stable
	UtilizationPct    string
	LoadBarPct        float64 // 0–100 for meter width (capped)
	MatchOK           bool
	MatchNote         string
	Receipt           string
	ParamRows         [][]string
	AdjustmentBullets []string
	NoBaseFee         bool

	StatBaseFee   string
	StatGasPrice  string
	StatLastLoad  string
	StatNextAdj   string
	WalletGasPrice string
	WalletPayNote  string
	ChainBaseFee   string
	ChainNoBaseFee string
	ChainDemandNote string
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

func formatUint(n uint64) string {
	if n == ^uint64(0) {
		return "unlimited"
	}
	return report.FormatInt(int64(n))
}

func formatDecAmount(d math.LegacyDec, denom string) string {
	if d.IsNil() {
		return "—"
	}
	return fetch.FormatFeeAmount(d.String(), denom)
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

func feemarketTraffic(wanted, target uint64, hasTarget bool) (label, class, nextAdj string) {
	if !hasTarget {
		return "FEE ADJUST UNKNOWN", "stable", "?"
	}
	switch {
	case wanted > target:
		return "FEE RISING", "rising", "↑"
	case wanted < target:
		return "FEE FALLING", "falling", "↓"
	default:
		return "STABLE", "stable", "="
	}
}

func buildFeemarketParamRows(d model.Report, target uint64, hasTarget bool) [][]string {
	var rows [][]string
	if d.Elasticity > 0 {
		val := fmt.Sprintf("%d", d.Elasticity)
		desc := "Target block fullness = max block gas ÷ elasticity (EIP-1559 style half-full aim)"
		if hasTarget {
			val += fmt.Sprintf(" → target %s gas", formatUint(target))
		}
		rows = append(rows, []string{"elasticity_multiplier", val, desc})
	}
	if d.BlockGasLimit > 0 {
		rows = append(rows, []string{
			"max block gas (consensus)",
			formatUint(d.BlockGasLimit),
			"Hard cap on gas per block from CometBFT consensus params",
		})
	}
	if d.BaseFeeChangeDenominator > 0 {
		rows = append(rows, []string{
			"base_fee_change_denominator",
			fmt.Sprintf("%d", d.BaseFeeChangeDenominator),
			"Larger value → slower base-fee moves per block (max change ≈ 1/d per block)",
		})
	}
	if d.MinGasPrice != "" {
		rows = append(rows, []string{
			"min_gas_price (floor)",
			d.MinGasPrice,
			"Base fee will not drop below this chain minimum",
		})
	}
	if d.MinGasMultiplier != "" {
		rows = append(rows, []string{
			"min_gas_multiplier",
			d.MinGasMultiplier,
			"EndBlock records demand as max(mempool gas × μ, gas actually used)",
		})
	}
	noBase := report.BoolStr(d.NoBaseFee)
	desc := "EIP-1559 auto-adjust active"
	if d.NoBaseFee {
		desc = "Fixed minimum gas — base fee does not move with congestion"
	}
	rows = append(rows, []string{"no_base_fee", noBase, desc})
	if d.AdjCap != "" {
		rows = append(rows, []string{"base_fee max change (cap)", d.AdjCap, "Upper bound on per-block base-fee delta when configured"})
	}
	return rows
}

func defaultAdjustmentBullets() []string {
	return []string{
		"After each block, the chain stores how busy it was (recorded demand).",
		"Compare recorded demand to the target capacity (max gas ÷ elasticity).",
		"Busier than target → raise minimum gas price next block; quieter → lower (not below floor).",
		"Wallets quote eth_gasPrice ≈ base fee + typical priority tip; you pay gas × (base + tip).",
	}
}

func buildFeemarketExplain(d model.Report) FeemarketExplain {
	ex := FeemarketExplain{}
	denom := diagramDenom(d)

	ex.StatBaseFee = d.BaseFee
	if d.GasPrice != "" {
		ex.StatGasPrice = fetch.FormatFeeAmount(d.GasPrice, denom)
		ex.WalletGasPrice = ex.StatGasPrice
	} else {
		ex.StatGasPrice = "—"
		ex.WalletGasPrice = "—"
	}
	ex.ChainBaseFee = d.BaseFee
	if d.BaseFee == "" {
		ex.ChainBaseFee = "—"
	}
	ex.ChainNoBaseFee = report.BoolStr(d.NoBaseFee)
	ex.WalletPayNote = "Typical cost ≈ gas used × (minimum gas price + optional priority tip)"
	ex.ChainDemandNote = "EndBlock demand from the previous block drives the base fee applied at BeginBlock of this block"

	if d.NoBaseFee {
		ex.NoBaseFee = true
		ex.SummaryLine = "Fixed gas pricing — EIP-1559 auto-adjust is off (no_base_fee)"
		ex.Hint = "Fees stay at the configured minimum; block congestion does not nudge the minimum gas price."
		ex.TrafficLabel = "FIXED PRICING"
		ex.TrafficClass = "stable"
		ex.NextAdj = "—"
		ex.StatNextAdj = "—"
		ex.Verdict = "no auto-adjust"
		ex.StatLastLoad = "n/a"
		ex.ParamRows = buildFeemarketParamRows(d, 0, false)
		ex.AdjustmentBullets = []string{
			"no_base_fee is enabled — the chain does not run EIP-1559 style base-fee updates.",
			"Wallets still use eth_gasPrice / min gas price from node config.",
		}
		ex.Receipt = strings.Join([]string{
			"1. EIP-1559 auto-adjust: OFF (no_base_fee)",
			"2. Minimum gas price: " + ex.ChainBaseFee,
			"3. eth_gasPrice: " + ex.WalletGasPrice,
		}, "\n")
		return ex
	}

	wanted := feemarketStoredWanted(d)
	target, hasTarget := feemarketGasTarget(d)
	gasUsed := d.ParentBlockGasUsed

	ex.TrafficLabel, ex.TrafficClass, ex.NextAdj = feemarketTraffic(wanted, target, hasTarget)
	ex.StatNextAdj = ex.NextAdj
	ex.Verdict = feemarketVerdictPlain(wanted, target, hasTarget)

	if hasTarget && target > 0 {
		pct := float64(wanted) / float64(target) * 100
		ex.UtilizationPct = fmt.Sprintf("%.1f%%", pct)
		ex.StatLastLoad = ex.UtilizationPct + " of target"
		ex.LoadBarPct = pct
		if ex.LoadBarPct > 100 {
			ex.LoadBarPct = 100
		}
	} else {
		ex.StatLastLoad = "unknown"
	}

	ex.SummaryLine = fmt.Sprintf("Block %s — base fee %s", d.BlockHeight, d.BaseFee)
	if ex.StatGasPrice != "—" {
		ex.SummaryLine += fmt.Sprintf(" · eth_gasPrice %s", ex.StatGasPrice)
	}
	if ex.UtilizationPct != "" {
		ex.SummaryLine += fmt.Sprintf(" · last block load %s of target", ex.UtilizationPct)
	}

	ex.Hint = "The **previous** block's load (recorded demand vs target) nudges the **minimum gas price** for this block — like a thermostat reacting to yesterday's temperature."

	ex.ParamRows = buildFeemarketParamRows(d, target, hasTarget)
	ex.AdjustmentBullets = defaultAdjustmentBullets()

	var receipt []string
	step := 1
	if !d.ParentBlockResultsOK {
		receipt = append(receipt, "⚠ Previous-block gas may be incomplete (block_results unavailable)")
	}
	receipt = append(receipt, fmt.Sprintf("%d. Previous block gas used: %s", step, formatUint(gasUsed)))
	step++

	demandLine := fmt.Sprintf("%d. Recorded demand (for next fee): %s", step, formatUint(wanted))
	if d.MinGasMultiplier != "" && d.MinGasMultiplier != "1" {
		demandLine += fmt.Sprintf("  [= max(mempool × %s, gas used)]", d.MinGasMultiplier)
	} else if wanted > 0 && wanted != gasUsed {
		demandLine += "  [= max(mempool, gas used)]"
	}
	receipt = append(receipt, demandLine)
	step++

	if hasTarget {
		receipt = append(receipt, fmt.Sprintf("%d. Target capacity: %s gas  (max %s ÷ elasticity %d)",
			step, formatUint(target), formatUint(d.BlockGasLimit), d.Elasticity))
		step++
		loadLine := fmt.Sprintf("%d. Load vs target: %s / %s", step, formatUint(wanted), formatUint(target))
		if ex.UtilizationPct != "" {
			loadLine += " ≈ " + ex.UtilizationPct
		}
		loadLine += " → " + ex.Verdict
		receipt = append(receipt, loadLine)
		step++
	} else {
		receipt = append(receipt, fmt.Sprintf("%d. Target capacity: unknown (need max block gas)", step))
		step++
	}

	receipt = append(receipt, fmt.Sprintf("%d. Base fee this block: %s", step, d.BaseFee))
	step++
	if ex.StatGasPrice != "—" {
		receipt = append(receipt, fmt.Sprintf("%d. eth_gasPrice: %s", step, ex.StatGasPrice))
		step++
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
			ex.MatchOK = calcGasBaseFee(wanted, target, denomU, parent, minUnit, minGasPrice).Equal(current)
		}
	}
	if ex.MatchOK {
		ex.MatchNote = "✓ recomputed base fee matches chain"
		receipt = append(receipt, fmt.Sprintf("%d. %s", step, ex.MatchNote))
	} else if okCurrent && hasTarget {
		ex.MatchNote = "⚠ could not verify base fee against parent (check block_results / params)"
		receipt = append(receipt, fmt.Sprintf("%d. %s", step, ex.MatchNote))
	}

	ex.Receipt = strings.Join(receipt, "\n")
	return ex
}
