package feemarket

import (
	"strings"

	"cosmossdk.io/math"
)

// MinUnitGas is the minimum base-fee increase step (1 apmt).
var MinUnitGas = math.LegacyOneDec()

// MaxUint64Label matches x/feemarket when consensus max_gas = -1.
const MaxUint64Label = "MaxUint64"

// StandardTransferGas is the gas limit for a simple EVM transfer.
const StandardTransferGas = 21_000

func CalcGasBaseFee(gasUsed, gasTarget, baseFeeChangeDenom uint64, baseFee, minUnitGas, minGasPrice math.LegacyDec) math.LegacyDec {
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

func ParseLegacyDec(s string) (math.LegacyDec, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return math.LegacyDec{}, false
	}
	d, err := math.LegacyNewDecFromStr(s)
	return d, err == nil
}

func InferParentBaseFee(current math.LegacyDec, gasUsed, gasTarget, denom uint64, minUnit, minGasPrice math.LegacyDec) (math.LegacyDec, bool) {
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
		got := CalcGasBaseFee(gasUsed, gasTarget, denom, parent, minUnit, minGasPrice)
		if got.Equal(current) {
			return parent, true
		}
	}
	return math.LegacyDec{}, false
}

// DecreaseStep returns |Δbase| when demand is below target (zero if increase path).
func DecreaseStep(wanted, target, denom uint64, baseFee math.LegacyDec) math.LegacyDec {
	if baseFee.IsNil() || target == 0 || denom == 0 || wanted >= target {
		return math.LegacyZeroDec()
	}
	num := math.LegacyNewDecFromInt(math.NewIntFromUint64(target).Sub(math.NewIntFromUint64(wanted)))
	num = num.Mul(baseFee)
	num = num.QuoInt(math.NewIntFromUint64(target))
	num = num.QuoInt(math.NewIntFromUint64(denom))
	return num
}

func VerifyMatch(current math.LegacyDec, wanted, target, denom uint64, minGasPrice math.LegacyDec) string {
	if target == 0 || current.IsNil() {
		return ""
	}
	parent, okParent := InferParentBaseFee(current, wanted, target, denom, MinUnitGas, minGasPrice)
	if okParent && CalcGasBaseFee(wanted, target, denom, parent, MinUnitGas, minGasPrice).Equal(current) {
		return "✓ match"
	}
	if okParent {
		return "⚠ inconclusive"
	}
	return ""
}
