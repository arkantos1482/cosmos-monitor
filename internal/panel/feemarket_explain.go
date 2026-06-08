package panel

import (
	"fmt"
	"strconv"
	"strings"

	"cosmossdk.io/math"
	"github.com/arkantos1482/cosmos-monitor/internal/fetch"
	"github.com/arkantos1482/cosmos-monitor/internal/model"
	"github.com/arkantos1482/cosmos-monitor/internal/report"
)

func diagramDenom(d model.Report) string {
	if d.EVMDenom != "" {
		return d.EVMDenom
	}
	if d.BondDenom != "" {
		return d.BondDenom
	}
	return "apmt"
}

func parseDiagramUint(s string) uint64 {
	s = strings.ReplaceAll(s, ",", "")
	n, _ := strconv.ParseUint(s, 10, 64)
	return n
}

// FeemarketExplain holds structured fee-market dashboard data (no LaTeX / Mermaid).
type FeemarketExplain struct {
	HeroLine          string
	TrafficLabel      string
	TrafficClass      string
	UtilizationPct    string
	LoadBarPct        float64
	HideLoadMeter     bool
	NextAdj           string
	VariableRows      [][]string
	FormulaBlocks     []string
	ParamRows         [][]string
	NoBaseFee         bool
	UnlimitedBlockGas bool
}

// feemarketMaxUint64Label matches x/feemarket when consensus max_gas = -1.
const feemarketMaxUint64Label = "MaxUint64"

type feemarketLoad struct {
	wanted            uint64
	target            uint64
	hasTarget         bool
	gasUsed           uint64
	utilPct           string
	loadBarPct        float64
	unlimitedBlockGas bool
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

func feemarketUnlimitedBlockGas(d model.Report) bool {
	return d.BlockGasLimit == ^uint64(0)
}

func formatUint(n uint64) string {
	if n == ^uint64(0) {
		return feemarketMaxUint64Label
	}
	return report.FormatInt(int64(n))
}

func feemarketLoadContext(d model.Report) feemarketLoad {
	ld := feemarketLoad{
		wanted:            feemarketStoredWanted(d),
		gasUsed:           d.ParentBlockGasUsed,
		unlimitedBlockGas: feemarketUnlimitedBlockGas(d),
	}
	ld.target, ld.hasTarget = feemarketGasTarget(d)
	if ld.unlimitedBlockGas && ld.hasTarget {
		return ld
	}
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

func feemarketCompareArrow(wanted, target uint64, hasTarget bool) string {
	if !hasTarget {
		return "?"
	}
	switch {
	case wanted > target:
		return "↑"
	case wanted < target:
		return "↓"
	default:
		return "="
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
		return "FEE STABLE", "stable", "="
	}
}

func feemarketHeroLine(d model.Report, noBaseFee bool) string {
	if noBaseFee {
		return fmt.Sprintf("Block %s · no_base_fee", d.BlockHeight)
	}
	return fmt.Sprintf("Block %s", d.BlockHeight)
}

func feemarketWMeaning(d model.Report) string {
	if d.MinGasMultiplier != "" && d.MinGasMultiplier != "1" {
		return fmt.Sprintf("Stored gas wanted (fee input); max(mempool × %s, gas_used)", d.MinGasMultiplier)
	}
	return "Stored gas wanted (fee input); max(mempool × min_gas_multiplier, gas_used)"
}

func feemarketTargetLiveValue(d model.Report, ld feemarketLoad) string {
	if !ld.hasTarget {
		return "—"
	}
	if ld.unlimitedBlockGas {
		return fmt.Sprintf("%s ÷ %d (sentinel)", feemarketMaxUint64Label, d.Elasticity)
	}
	return formatUint(ld.target)
}

func feemarketTargetMeaning(d model.Report, ld feemarketLoad) string {
	if ld.unlimitedBlockGas {
		return "gasLimit ÷ elasticity; max_gas = −1 → gasLimit = MaxUint64"
	}
	return "max_block_gas ÷ elasticity_multiplier"
}

func buildFeemarketParamRows(d model.Report) [][]string {
	var rows [][]string
	if d.Elasticity > 0 {
		meaning := "Target = gasLimit ÷ elasticity"
		if feemarketUnlimitedBlockGas(d) {
			meaning = "With max_gas = −1, target = MaxUint64 ÷ elasticity (sentinel)"
		}
		rows = append(rows, []string{
			"elasticity_multiplier",
			fmt.Sprintf("%d", d.Elasticity),
			meaning,
		})
	}
	if d.BlockGasLimit > 0 {
		val := formatUint(d.BlockGasLimit)
		meaning := "Consensus max_gas per block"
		if feemarketUnlimitedBlockGas(d) {
			val = "unlimited (max_gas = −1)"
			meaning = "Keeper uses MaxUint64 as gasLimit in CalculateBaseFee"
		}
		rows = append(rows, []string{
			"max block gas (consensus)",
			val,
			meaning,
		})
	}
	if d.BaseFeeChangeDenominator > 0 {
		rows = append(rows, []string{
			"base_fee_change_denominator",
			fmt.Sprintf("%d", d.BaseFeeChangeDenominator),
			"Caps per-block |Δbase| in adjustment formula",
		})
	}
	if d.MinGasPrice != "" {
		rows = append(rows, []string{
			"min_gas_price (floor)",
			d.MinGasPrice,
			"Base fee floor on decrease",
		})
	}
	if d.MinGasMultiplier != "" {
		rows = append(rows, []string{
			"min_gas_multiplier",
			d.MinGasMultiplier,
			"mempool gas × multiplier vs gas_used → stored W",
		})
	}
	noBase := report.BoolStr(d.NoBaseFee)
	note := "EIP-1559 auto-adjust enabled"
	if d.NoBaseFee {
		note = "Fixed minimum gas; congestion ignored"
	}
	rows = append(rows, []string{"no_base_fee", noBase, note})
	if d.AdjCap != "" {
		rows = append(rows, []string{
			"base_fee max change (cap)",
			d.AdjCap,
			"Optional upper bound on per-block base fee delta",
		})
	}
	return rows
}

func buildFeemarketVariableRows(d model.Report, ld feemarketLoad) [][]string {
	var rows [][]string
	rows = append(rows, []string{"W", feemarketWMeaning(d), formatUint(ld.wanted)})
	rows = append(rows, []string{"gas_used", "Gas consumed in block N−1", formatUint(ld.gasUsed)})
	if ld.hasTarget || ld.unlimitedBlockGas {
		rows = append(rows, []string{
			"target",
			feemarketTargetMeaning(d, ld),
			feemarketTargetLiveValue(d, ld),
		})
	}
	if d.BlockGasLimit > 0 {
		gasLimitVal := formatUint(d.BlockGasLimit)
		gasLimitMeaning := "Consensus max_block_gas (gasLimit in keeper)"
		if ld.unlimitedBlockGas {
			gasLimitVal = "unlimited (−1 → MaxUint64)"
			gasLimitMeaning = "Consensus max_gas = −1; keeper gasLimit = MaxUint64"
		}
		rows = append(rows, []string{"gasLimit", gasLimitMeaning, gasLimitVal})
	}
	baseLabel := "base"
	baseMeaning := "Base fee this block (BeginBlock)"
	if d.NoBaseFee {
		baseLabel = "base (fixed)"
		baseMeaning = "Fixed base fee (no_base_fee)"
	}
	baseVal := d.BaseFee
	if baseVal == "" {
		baseVal = "—"
	}
	rows = append(rows, []string{baseLabel, baseMeaning, baseVal})
	if d.BaseFeeChangeDenominator > 0 && !d.NoBaseFee {
		rows = append(rows, []string{
			"denom",
			"base_fee_change_denominator",
			fmt.Sprintf("%d", d.BaseFeeChangeDenominator),
		})
	}
	if gp := formatGasPriceStat(d); gp != "—" {
		rows = append(rows, []string{"eth_gasPrice", "EVM JSON-RPC gas price quote", gp})
	}
	if !d.NoBaseFee && ld.hasTarget && d.BaseFeeChangeDenominator > 0 {
		if verify := feemarketVerifyRow(d, ld); verify != "" {
			rows = append(rows, []string{"verify", "Recomputed base fee vs chain", verify})
		}
	}
	return rows
}

func feemarketVerifyRow(d model.Report, ld feemarketLoad) string {
	current, okCurrent := parseLegacyDec(d.BaseFeeRaw)
	if !okCurrent {
		return ""
	}
	minGasPrice := math.LegacyZeroDec()
	if mp, ok := parseLegacyDec(d.MinGasPriceRaw); ok {
		minGasPrice = mp
	}
	denomU := uint64(d.BaseFeeChangeDenominator)
	parent, okParent := inferParentBaseFee(current, ld.wanted, ld.target, denomU, math.LegacyOneDec(), minGasPrice)
	if okParent && calcGasBaseFee(ld.wanted, ld.target, denomU, parent, math.LegacyOneDec(), minGasPrice).Equal(current) {
		return "✓ match"
	}
	if okParent {
		return "⚠ inconclusive"
	}
	return ""
}

func buildFeemarketFormulaBlocks(d model.Report, ld feemarketLoad) []string {
	if d.NoBaseFee || d.BaseFeeChangeDenominator == 0 {
		return nil
	}
	base := d.BaseFee
	if base == "" {
		base = d.BaseFeeRaw
	}
	denom := d.BaseFeeChangeDenominator
	arrow := feemarketCompareArrow(ld.wanted, ld.target, ld.hasTarget)
	wStr := formatUint(ld.wanted)
	tStr := feemarketTargetLiveValue(d, ld)

	var blocks []string
	if ld.unlimitedBlockGas && d.Elasticity > 0 {
		blocks = append(blocks, fmt.Sprintf(
			"target = gasLimit ÷ elasticity\n"+
				"       = unlimited → %s ÷ %d\n"+
				"       = %s ÷ %d (sentinel)",
			feemarketMaxUint64Label, d.Elasticity,
			feemarketMaxUint64Label, d.Elasticity,
		))
	} else if ld.hasTarget && d.Elasticity > 0 && d.BlockGasLimit > 0 {
		blocks = append(blocks, fmt.Sprintf(
			"target = max_block_gas ÷ elasticity\n"+
				"       = %s ÷ %d\n"+
				"       = %s",
			formatUint(d.BlockGasLimit), d.Elasticity,
			formatUint(ld.target),
		))
	}

	if ld.hasTarget {
		blocks = append(blocks, fmt.Sprintf(
			"|Δbase| ≤ base × |W − target| / (target × denom)\n"+
				"        ≤ %s × |%s − %s| / (%s × %d)\n"+
				"        → %s",
			base, wStr, tStr, tStr, denom, arrow,
		))
	}
	return blocks
}

func formatGasPriceStat(d model.Report) string {
	denom := diagramDenom(d)
	if d.GasPrice != "" {
		return fetch.FormatFeeAmount(d.GasPrice, denom)
	}
	return "—"
}

func buildFeemarketExplain(d model.Report) FeemarketExplain {
	ex := FeemarketExplain{}
	ld := feemarketLoadContext(d)

	ex.NoBaseFee = d.NoBaseFee
	ex.ParamRows = buildFeemarketParamRows(d)
	ex.VariableRows = buildFeemarketVariableRows(d, ld)

	if d.NoBaseFee {
		ex.TrafficLabel = "FIXED PRICING"
		ex.TrafficClass = "stable"
		ex.NextAdj = "—"
		ex.HeroLine = feemarketHeroLine(d, true)
		return ex
	}

	ex.UnlimitedBlockGas = ld.unlimitedBlockGas
	ex.HideLoadMeter = ld.unlimitedBlockGas
	ex.TrafficLabel, ex.TrafficClass, ex.NextAdj = feemarketTraffic(ld.wanted, ld.target, ld.hasTarget)
	ex.UtilizationPct = ld.utilPct
	ex.LoadBarPct = ld.loadBarPct
	ex.HeroLine = feemarketHeroLine(d, false)
	ex.FormulaBlocks = buildFeemarketFormulaBlocks(d, ld)
	return ex
}
