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
	HeroLine       string
	TrafficLabel   string
	TrafficClass   string
	UtilizationPct string
	LoadBarPct     float64
	NextAdj        string
	LastBlockRows  [][]string
	FormulaLine    string
	ThisBlockRows  [][]string
	ParamRows      [][]string
	WalletLine     string
	ChainLine      string
	NoBaseFee      bool
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
		rows = append(rows, []string{
			"elasticity_multiplier",
			fmt.Sprintf("%d", d.Elasticity),
			"Target gas = max_block_gas ÷ elasticity",
		})
	}
	if d.BlockGasLimit > 0 {
		rows = append(rows, []string{
			"max block gas (consensus)",
			formatUint(d.BlockGasLimit),
			"Hard cap per block (consensus max_gas)",
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
		targetVal := formatUint(ld.target)
		if d.BlockGasLimit > 0 && d.Elasticity > 0 {
			targetVal += fmt.Sprintf(" (= %s ÷ %d)", formatUint(d.BlockGasLimit), d.Elasticity)
		}
		rows = append(rows, []string{
			"target",
			targetVal,
			"`GET /cosmos/evm/feemarket/v1/params`",
		})
		arrow := feemarketCompareArrow(ld.wanted, ld.target, true)
		rows = append(rows, []string{
			"W vs target",
			fmt.Sprintf("%s vs %s → %s", formatUint(ld.wanted), formatUint(ld.target), arrow),
			"derived",
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

func buildFeemarketFormulaLine(d model.Report) string {
	if d.NoBaseFee || d.BaseFeeChangeDenominator == 0 {
		return ""
	}
	line := fmt.Sprintf(
		"Δbase ≤ base × |W − target| / (target × %d)",
		d.BaseFeeChangeDenominator,
	)
	var extras []string
	if d.MinGasPrice != "" {
		extras = append(extras, "min_gas_price floor on decrease")
	}
	if d.MinGasMultiplier != "" {
		extras = append(extras, "W from min_gas_multiplier × mempool vs gas_used")
	}
	if len(extras) > 0 {
		line += "; " + strings.Join(extras, "; ")
	}
	return line
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

	ex.TrafficLabel, ex.TrafficClass, ex.NextAdj = feemarketTraffic(ld.wanted, ld.target, ld.hasTarget)
	ex.UtilizationPct = ld.utilPct
	ex.LoadBarPct = ld.loadBarPct
	ex.HeroLine = feemarketHeroLine(d, ld)
	ex.FormulaLine = buildFeemarketFormulaLine(d)
	return ex
}
