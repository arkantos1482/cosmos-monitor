package panel

import (
	"fmt"
	"strings"

	"cosmossdk.io/math"
	"github.com/arkantos1482/cosmos-monitor/internal/fetch"
	"github.com/arkantos1482/cosmos-monitor/internal/model"
	"github.com/arkantos1482/cosmos-monitor/internal/report"
)

// FeemarketExplain holds structured fee-market dashboard data (no LaTeX / Mermaid).
type FeemarketExplain struct {
	HeroLine         string
	TrafficLabel     string
	TrafficClass     string
	UtilizationPct   string
	LoadBarPct       float64
	NextAdj          string
	StatBaseFee      string
	StatGasPrice     string
	StatNextAdj      string
	ParamRows        [][]string
	Receipt          string
	WalletLine       string
	ChainLine        string
	AdjustmentBullets []string // 0–2 live bullets; omit when receipt/table suffice
	NoBaseFee        bool
}

type feemarketLoad struct {
	wanted     uint64
	target     uint64
	hasTarget  bool
	gasUsed    uint64
	utilPct    string
	loadBarPct float64
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

func feemarketLoadContext(d model.Report) feemarketLoad {
	ld := feemarketLoad{
		wanted:  feemarketStoredWanted(d),
		gasUsed: d.ParentBlockGasUsed,
	}
	ld.target, ld.hasTarget = feemarketGasTarget(d)
	if ld.hasTarget && ld.target > 0 {
		pct := float64(ld.wanted) / float64(ld.target) * 100
		ld.utilPct = fmt.Sprintf("%.2f%%", pct)
		ld.loadBarPct = pct
		if ld.loadBarPct > 100 {
			ld.loadBarPct = 100
		}
	}
	return ld
}

func feemarketVerdictShort(wanted, target uint64, hasTarget bool) string {
	if !hasTarget {
		return "fee direction unknown"
	}
	switch {
	case wanted > target:
		return "fee rising"
	case wanted < target:
		return "fee falling"
	default:
		return "fee stable"
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

func feemarketHeroLine(d model.Report, ld feemarketLoad, verdict string) string {
	line := fmt.Sprintf("Block %s", d.BlockHeight)
	if ld.hasTarget {
		line += fmt.Sprintf(" · demand %s / target %s (%s) → %s",
			formatUint(ld.wanted), formatUint(ld.target), ld.utilPct, verdict)
	} else {
		line += fmt.Sprintf(" · demand %s → %s", formatUint(ld.wanted), verdict)
	}
	return line
}

func feemarketTargetFormula(d model.Report, target uint64) string {
	if d.BlockGasLimit > 0 && d.Elasticity > 0 {
		return fmt.Sprintf("%s ÷ elasticity %d = %s gas",
			formatUint(d.BlockGasLimit), d.Elasticity, formatUint(target))
	}
	return "unknown (need max block gas and elasticity)"
}

func buildFeemarketParamRows(d model.Report, ld feemarketLoad) [][]string {
	var rows [][]string
	if d.Elasticity > 0 {
		note := feemarketTargetFormula(d, ld.target)
		if ld.hasTarget {
			note += fmt.Sprintf(" — last block %s of target (%s)", ld.utilPct, feemarketVerdictShort(ld.wanted, ld.target, true))
		}
		rows = append(rows, []string{
			"elasticity_multiplier",
			fmt.Sprintf("%d", d.Elasticity),
			note,
		})
	}
	if d.BlockGasLimit > 0 {
		rows = append(rows, []string{
			"max block gas (consensus)",
			formatUint(d.BlockGasLimit),
			fmt.Sprintf("Hard cap per block; target gas uses %s", feemarketTargetFormula(d, ld.target)),
		})
	}
	if d.BaseFeeChangeDenominator > 0 {
		note := fmt.Sprintf("Per-block |Δbase| ≤ base × |demand−target| / (target × %d)", d.BaseFeeChangeDenominator)
		if ld.hasTarget && d.BaseFee != "" {
			note += fmt.Sprintf("; at current load (%s) next block moves %s", ld.utilPct, feemarketVerdictShort(ld.wanted, ld.target, true))
		}
		rows = append(rows, []string{
			"base_fee_change_denominator",
			fmt.Sprintf("%d", d.BaseFeeChangeDenominator),
			note,
		})
	}
	if d.MinGasPrice != "" {
		rows = append(rows, []string{
			"min_gas_price (floor)",
			d.MinGasPrice,
			fmt.Sprintf("Base fee will not fall below %s on decrease", d.MinGasPrice),
		})
	}
	if d.MinGasMultiplier != "" {
		rows = append(rows, []string{
			"min_gas_multiplier",
			d.MinGasMultiplier,
			fmt.Sprintf("Recorded demand = max(mempool gas × %s, gas used %s)", d.MinGasMultiplier, formatUint(ld.gasUsed)),
		})
	}
	noBase := report.BoolStr(d.NoBaseFee)
	note := "EIP-1559 auto-adjust on — base fee moves with congestion"
	if d.NoBaseFee {
		note = "Fixed minimum gas — congestion does not move base fee"
	}
	rows = append(rows, []string{"no_base_fee", noBase, note})
	if d.AdjCap != "" {
		rows = append(rows, []string{
			"base_fee max change (cap)",
			d.AdjCap,
			fmt.Sprintf("Upper bound on per-block delta when set (denominator %d)", d.BaseFeeChangeDenominator),
		})
	}
	return rows
}

func feemarketWalletLine(d model.Report, gasPrice string) string {
	if d.NoBaseFee {
		return fmt.Sprintf("eth_gasPrice %s — fixed pricing (no_base_fee=true); pay ≈ gas × quoted price", gasPrice)
	}
	return fmt.Sprintf("eth_gasPrice %s — pay ≈ gas × (base %s + priority tip)", gasPrice, d.BaseFee)
}

func feemarketChainLine(d model.Report, ld feemarketLoad) string {
	if d.NoBaseFee {
		return fmt.Sprintf("Chain enforces min gas %s; no_base_fee=true — prior block load does not adjust fees", d.BaseFee)
	}
	if ld.hasTarget {
		return fmt.Sprintf("BeginBlock base fee %s from prior demand %s vs target %s (%s, %s)",
			d.BaseFee, formatUint(ld.wanted), formatUint(ld.target), ld.utilPct,
			feemarketVerdictShort(ld.wanted, ld.target, true))
	}
	return fmt.Sprintf("BeginBlock base fee %s — target unknown (need max gas ÷ elasticity)", d.BaseFee)
}

func feemarketAdjustmentBullets(d model.Report, ld feemarketLoad, nextAdj string) []string {
	if d.NoBaseFee {
		return []string{
			fmt.Sprintf("no_base_fee=true — minimum stays at %s; recorded demand %s is informational only",
				d.BaseFee, formatUint(ld.wanted)),
			fmt.Sprintf("Wallets quote eth_gasPrice %s from node JSON-RPC", formatGasPriceStat(d)),
		}
	}
	if !ld.hasTarget {
		return nil
	}
	return []string{
		fmt.Sprintf("Prior block demand %s vs target %s (%s) → next minimum gas price %s (%s)",
			formatUint(ld.wanted), formatUint(ld.target), ld.utilPct, nextAdj,
			feemarketVerdictShort(ld.wanted, ld.target, true)),
		fmt.Sprintf("Max block %s ÷ elasticity %d = target %s; denominator %d caps per-block %s moves",
			formatUint(d.BlockGasLimit), d.Elasticity, formatUint(ld.target),
			d.BaseFeeChangeDenominator, nextAdj),
	}
}

func formatGasPriceStat(d model.Report) string {
	denom := diagramDenom(d)
	if d.GasPrice != "" {
		return fetch.FormatFeeAmount(d.GasPrice, denom)
	}
	return "—"
}

func buildFeemarketReceipt(d model.Report, ld feemarketLoad, ex *FeemarketExplain) string {
	if d.NoBaseFee {
		return strings.Join([]string{
			"1. EIP-1559 auto-adjust: OFF (no_base_fee=true)",
			"2. Minimum gas price: " + ex.StatBaseFee,
			"3. eth_gasPrice: " + ex.StatGasPrice,
		}, "\n")
	}

	var lines []string
	step := 1
	if !d.ParentBlockResultsOK {
		lines = append(lines, "⚠ Previous-block gas may be incomplete (block_results unavailable)")
	}
	lines = append(lines, fmt.Sprintf("%d. Previous block gas used: %s", step, formatUint(ld.gasUsed)))
	step++

	demandLine := fmt.Sprintf("%d. Recorded demand (fee input): %s", step, formatUint(ld.wanted))
	if d.MinGasMultiplier != "" && d.MinGasMultiplier != "1" {
		demandLine += fmt.Sprintf(" [= max(mempool × %s, gas used)]", d.MinGasMultiplier)
	} else if ld.wanted > 0 && ld.wanted != ld.gasUsed {
		demandLine += " [= max(mempool, gas used)]"
	}
	lines = append(lines, demandLine)
	step++

	if ld.hasTarget {
		lines = append(lines, fmt.Sprintf("%d. Target: %s  (%s)",
			step, formatUint(ld.target), feemarketTargetFormula(d, ld.target)))
		step++
		loadLine := fmt.Sprintf("%d. Load: %s / %s", step, formatUint(ld.wanted), formatUint(ld.target))
		if ld.utilPct != "" {
			loadLine += " = " + ld.utilPct
		}
		loadLine += " → " + feemarketVerdictShort(ld.wanted, ld.target, true) + " (" + ex.NextAdj + " next block)"
		lines = append(lines, loadLine)
		step++
	} else {
		lines = append(lines, fmt.Sprintf("%d. Target: unknown (need max block gas + elasticity)", step))
		step++
	}

	lines = append(lines, fmt.Sprintf("%d. Base fee this block: %s", step, d.BaseFee))
	step++
	if ex.StatGasPrice != "—" {
		lines = append(lines, fmt.Sprintf("%d. eth_gasPrice: %s", step, ex.StatGasPrice))
		step++
	}

	current, okCurrent := parseLegacyDec(d.BaseFeeRaw)
	if okCurrent && ld.hasTarget && d.BaseFeeChangeDenominator > 0 {
		minGasPrice := math.LegacyZeroDec()
		if mp, ok := parseLegacyDec(d.MinGasPriceRaw); ok {
			minGasPrice = mp
		}
		denomU := uint64(d.BaseFeeChangeDenominator)
		parent, okParent := inferParentBaseFee(current, ld.wanted, ld.target, denomU, math.LegacyOneDec(), minGasPrice)
		if okParent && calcGasBaseFee(ld.wanted, ld.target, denomU, parent, math.LegacyOneDec(), minGasPrice).Equal(current) {
			lines = append(lines, fmt.Sprintf("%d. ✓ recomputed base fee matches chain (parent → current)", step))
		} else if okParent {
			lines = append(lines, fmt.Sprintf("%d. ⚠ base fee verify inconclusive (check block_results / params)", step))
		}
	}
	return strings.Join(lines, "\n")
}

func buildFeemarketExplain(d model.Report) FeemarketExplain {
	ex := FeemarketExplain{}
	ld := feemarketLoadContext(d)

	ex.StatBaseFee = d.BaseFee
	if ex.StatBaseFee == "" {
		ex.StatBaseFee = "—"
	}
	ex.StatGasPrice = formatGasPriceStat(d)
	ex.NoBaseFee = d.NoBaseFee

	if d.NoBaseFee {
		ex.TrafficLabel = "FIXED PRICING"
		ex.TrafficClass = "stable"
		ex.NextAdj = "—"
		ex.StatNextAdj = "—"
		ex.HeroLine = fmt.Sprintf("Block %s · fixed gas pricing (no_base_fee) · base %s", d.BlockHeight, ex.StatBaseFee)
		ex.ParamRows = buildFeemarketParamRows(d, ld)
		ex.WalletLine = feemarketWalletLine(d, ex.StatGasPrice)
		ex.ChainLine = feemarketChainLine(d, ld)
		ex.AdjustmentBullets = feemarketAdjustmentBullets(d, ld, ex.NextAdj)
		ex.Receipt = buildFeemarketReceipt(d, ld, &ex)
		return ex
	}

	verdict := feemarketVerdictShort(ld.wanted, ld.target, ld.hasTarget)
	ex.TrafficLabel, ex.TrafficClass, ex.NextAdj = feemarketTraffic(ld.wanted, ld.target, ld.hasTarget)
	ex.StatNextAdj = ex.NextAdj
	ex.UtilizationPct = ld.utilPct
	ex.LoadBarPct = ld.loadBarPct
	ex.HeroLine = feemarketHeroLine(d, ld, verdict)
	ex.ParamRows = buildFeemarketParamRows(d, ld)
	ex.WalletLine = feemarketWalletLine(d, ex.StatGasPrice)
	ex.ChainLine = feemarketChainLine(d, ld)
	ex.AdjustmentBullets = feemarketAdjustmentBullets(d, ld, ex.NextAdj)
	ex.Receipt = buildFeemarketReceipt(d, ld, &ex)
	return ex
}
