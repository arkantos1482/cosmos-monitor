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
	HeroLine           string
	TrafficLabel       string
	TrafficClass       string
	UtilizationPct     string
	LoadBarPct         float64
	HideLoadMeter      bool
	FormulaHeading     string
	NextAdj            string
	LastBlockRows      [][]string
	FormulaLine        string
	ThisBlockRows      [][]string
	ParamRows          [][]string
	WalletLine         string
	ChainLine          string
	NoBaseFee          bool
	UnlimitedBlockGas  bool
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
		ld.utilPct = "n/a — max_gas is unlimited (-1)"
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

func feemarketTargetDisplay(d model.Report) (value, source string) {
	if d.Elasticity <= 0 {
		return "—", "need elasticity_multiplier from feemarket params"
	}
	if feemarketUnlimitedBlockGas(d) {
		return fmt.Sprintf("%s ÷ %d (sentinel — not a real block gas budget)",
				feemarketMaxUint64Label, d.Elasticity),
			"`x/feemarket` CalculateBaseFee: gasLimit = MaxUint64 when consensus max_gas = -1"
	}
	if d.BlockGasLimit > 0 {
		t := d.BlockGasLimit / uint64(d.Elasticity)
		return fmt.Sprintf("%s (= %s ÷ %d)",
				formatUint(t), formatUint(d.BlockGasLimit), d.Elasticity),
			"max_block_gas ÷ elasticity_multiplier"
	}
	return "—", "need consensus max_block_gas + elasticity_multiplier"
}

func feemarketCompareDisplay(d model.Report, ld feemarketLoad) string {
	arrow := feemarketCompareArrow(ld.wanted, ld.target, ld.hasTarget)
	if !ld.hasTarget {
		return "?"
	}
	if ld.unlimitedBlockGas {
		return fmt.Sprintf("W %s ≪ %s ÷ %d → %s",
			formatUint(ld.wanted), feemarketMaxUint64Label, d.Elasticity, arrow)
	}
	return fmt.Sprintf("%s vs %s → %s", formatUint(ld.wanted), formatUint(ld.target), arrow)
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

func feemarketHeroLine(d model.Report, ld feemarketLoad) string {
	line := fmt.Sprintf("Block %s", d.BlockHeight)
	if ld.unlimitedBlockGas && ld.hasTarget {
		line += fmt.Sprintf(" · W %s · consensus max_gas unlimited (-1) · keeper target %s ÷ %d · %s",
			formatUint(ld.wanted), feemarketMaxUint64Label, d.Elasticity,
			feemarketVerdictShort(ld.wanted, ld.target, true))
		return line
	}
	if ld.hasTarget {
		line += fmt.Sprintf(" · W %s / target %s (%s)",
			formatUint(ld.wanted), formatUint(ld.target), ld.utilPct)
	} else if ld.wanted > 0 {
		line += fmt.Sprintf(" · W %s", formatUint(ld.wanted))
	}
	return line
}

func buildFeemarketParamRows(d model.Report) [][]string {
	var rows [][]string
	if d.Elasticity > 0 {
		meaning := "Target gas = max_block_gas ÷ elasticity"
		if feemarketUnlimitedBlockGas(d) {
			meaning = "With max_gas = -1, target = MaxUint64 ÷ elasticity (sentinel in keeper)"
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
			val = "unlimited (max_gas = -1)"
			meaning = "Keeper uses MaxUint64 as gasLimit in CalculateBaseFee (same as chain code)"
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
			"Caps per-block |Δbase| relative to |W − target| / target",
		})
	}
	if d.MinGasPrice != "" {
		rows = append(rows, []string{
			"min_gas_price (floor)",
			d.MinGasPrice,
			"Base fee will not fall below this on decrease",
		})
	}
	if d.MinGasMultiplier != "" {
		rows = append(rows, []string{
			"min_gas_multiplier",
			d.MinGasMultiplier,
			"Stored W = max(mempool gas × multiplier, block gas_used)",
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
			"Optional upper bound on per-block base fee delta",
		})
	}
	return rows
}

func feemarketLastBlockSourceResults(d model.Report) string {
	if d.ParentBlockResultsOK {
		return "CometBFT `block_results` (H−1)"
	}
	return "CometBFT `block_results` (H−1) ⚠ unavailable"
}

func buildFeemarketLastBlockRows(d model.Report, ld feemarketLoad) [][]string {
	var rows [][]string
	rows = append(rows, []string{
		"gas_used",
		formatUint(ld.gasUsed),
		feemarketLastBlockSourceResults(d),
	})

	wantedVal := formatUint(ld.wanted)
	if d.MinGasMultiplier != "" && d.MinGasMultiplier != "1" {
		wantedVal += fmt.Sprintf(" (max(mempool × %s, gas_used))", d.MinGasMultiplier)
	} else if ld.wanted > 0 && ld.wanted != ld.gasUsed {
		wantedVal += " (max(mempool, gas_used))"
	}
	rows = append(rows, []string{
		"gas_wanted (stored W)",
		wantedVal,
		"`GET /cosmos/evm/feemarket/v1/block_gas`",
	})

	if ld.hasTarget {
		targetVal, targetSrc := feemarketTargetDisplay(d)
		rows = append(rows, []string{
			"target",
			targetVal,
			targetSrc,
		})
		rows = append(rows, []string{
			"W vs target",
			feemarketCompareDisplay(d, ld),
			"derived from prior W and keeper target",
		})
	} else {
		rows = append(rows, []string{
			"target",
			"—",
			"need max_block_gas + elasticity_multiplier",
		})
	}
	return rows
}

func buildFeemarketFormulaLine(d model.Report, ld feemarketLoad) (heading, line string) {
	if d.NoBaseFee || d.BaseFeeChangeDenominator == 0 {
		return "", ""
	}
	heading = "How fees adjust"
	if ld.unlimitedBlockGas && d.Elasticity > 0 {
		line = fmt.Sprintf(
			"Consensus max_gas is -1 (unlimited). The keeper does not divide infinity: it uses gasLimit = %s, then target = %s ÷ elasticity %d. "+
				"That target is a sentinel for the EIP-1559 formula — real blocks never approach it. Prior block W = %s is far below it, so base fee tends to decrease each block (until min_gas_price). "+
				"Per-block change is capped by base_fee_change_denominator %d.",
			feemarketMaxUint64Label, feemarketMaxUint64Label, d.Elasticity,
			formatUint(ld.wanted), d.BaseFeeChangeDenominator,
		)
		return heading, line
	}
	if ld.hasTarget {
		line = fmt.Sprintf(
			"Target = max_block_gas ÷ elasticity = %s. Prior W = %s → %s next block. "+
				"Change bounded by base × |W−target| / (target × %d); min_gas_price caps decreases.",
			formatUint(ld.target), formatUint(ld.wanted),
			feemarketCompareArrow(ld.wanted, ld.target, true), d.BaseFeeChangeDenominator,
		)
	} else {
		line = fmt.Sprintf(
			"Change bounded by base × |W−target| / (target × %d); need max_block_gas and elasticity to compute target.",
			d.BaseFeeChangeDenominator,
		)
	}
	if d.MinGasMultiplier != "" {
		line += " W uses min_gas_multiplier × mempool vs gas_used."
	}
	return heading, line
}

func formatGasPriceStat(d model.Report) string {
	denom := diagramDenom(d)
	if d.GasPrice != "" {
		return fetch.FormatFeeAmount(d.GasPrice, denom)
	}
	return "—"
}

func buildFeemarketThisBlockRows(d model.Report, ld feemarketLoad) [][]string {
	baseFee := d.BaseFee
	if baseFee == "" {
		baseFee = "—"
	}
	var rows [][]string
	if d.NoBaseFee {
		rows = append(rows, []string{
			"base_fee (fixed)",
			baseFee,
			"`GET /cosmos/evm/feemarket/v1/base_fee` (no_base_fee)",
		})
	} else {
		rows = append(rows, []string{
			"base_fee",
			baseFee,
			"`GET /cosmos/evm/feemarket/v1/base_fee`",
		})
	}
	if gp := formatGasPriceStat(d); gp != "—" {
		rows = append(rows, []string{
			"eth_gasPrice",
			gp,
			"EVM JSON-RPC `eth_gasPrice`",
		})
	}
	if !d.NoBaseFee && ld.hasTarget && d.BaseFeeChangeDenominator > 0 {
		current, okCurrent := parseLegacyDec(d.BaseFeeRaw)
		if okCurrent {
			minGasPrice := math.LegacyZeroDec()
			if mp, ok := parseLegacyDec(d.MinGasPriceRaw); ok {
				minGasPrice = mp
			}
			denomU := uint64(d.BaseFeeChangeDenominator)
			parent, okParent := inferParentBaseFee(current, ld.wanted, ld.target, denomU, math.LegacyOneDec(), minGasPrice)
			if okParent && calcGasBaseFee(ld.wanted, ld.target, denomU, parent, math.LegacyOneDec(), minGasPrice).Equal(current) {
				rows = append(rows, []string{
					"verify",
					"✓ recomputed base fee matches chain",
					"inferParentBaseFee + calcGasBaseFee",
				})
			} else if okParent {
				rows = append(rows, []string{
					"verify",
					"⚠ inconclusive (check block_results / params)",
					"inferParentBaseFee + calcGasBaseFee",
				})
			}
		}
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
	if ld.unlimitedBlockGas && ld.hasTarget {
		return fmt.Sprintf("BeginBlock base fee %s: prior W %s vs keeper sentinel %s ÷ %d — W is far below → %s",
			d.BaseFee, formatUint(ld.wanted), feemarketMaxUint64Label, d.Elasticity,
			feemarketVerdictShort(ld.wanted, ld.target, true))
	}
	if ld.hasTarget {
		return fmt.Sprintf("BeginBlock base fee %s from prior W %s vs target %s (%s, %s)",
			d.BaseFee, formatUint(ld.wanted), formatUint(ld.target), ld.utilPct,
			feemarketVerdictShort(ld.wanted, ld.target, true))
	}
	return fmt.Sprintf("BeginBlock base fee %s — target unknown (need max gas ÷ elasticity)", d.BaseFee)
}

func buildFeemarketExplain(d model.Report) FeemarketExplain {
	ex := FeemarketExplain{}
	ld := feemarketLoadContext(d)
	gasPrice := formatGasPriceStat(d)

	ex.NoBaseFee = d.NoBaseFee
	ex.LastBlockRows = buildFeemarketLastBlockRows(d, ld)
	ex.ThisBlockRows = buildFeemarketThisBlockRows(d, ld)
	ex.ParamRows = buildFeemarketParamRows(d)
	ex.WalletLine = feemarketWalletLine(d, gasPrice)
	ex.ChainLine = feemarketChainLine(d, ld)

	if d.NoBaseFee {
		ex.TrafficLabel = "FIXED PRICING"
		ex.TrafficClass = "stable"
		ex.NextAdj = "—"
		ex.HeroLine = fmt.Sprintf("Block %s · fixed pricing (no_base_fee)", d.BlockHeight)
		return ex
	}

	ex.UnlimitedBlockGas = ld.unlimitedBlockGas
	ex.HideLoadMeter = ld.unlimitedBlockGas
	ex.TrafficLabel, ex.TrafficClass, ex.NextAdj = feemarketTraffic(ld.wanted, ld.target, ld.hasTarget)
	ex.UtilizationPct = ld.utilPct
	ex.LoadBarPct = ld.loadBarPct
	ex.HeroLine = feemarketHeroLine(d, ld)
	ex.FormulaHeading, ex.FormulaLine = buildFeemarketFormulaLine(d, ld)
	return ex
}
