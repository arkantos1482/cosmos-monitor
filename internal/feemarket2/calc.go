package feemarket2

import sdkmath "cosmossdk.io/math"

// MinUnitGas is the minimum base-fee increase step (1 apmt).
var MinUnitGas = sdkmath.LegacyOneDec()

// CalcGasBaseFee mirrors x/feemarket/types.CalcGasBaseFee (EIP-1559).
func CalcGasBaseFee(gasUsed, gasTarget, baseFeeChangeDenom uint64, baseFee, minUnitGas, minGasPrice sdkmath.LegacyDec) sdkmath.LegacyDec {
	if baseFee.IsNil() {
		return sdkmath.LegacyZeroDec()
	}
	if gasUsed == gasTarget {
		return baseFee
	}
	if gasTarget == 0 {
		return sdkmath.LegacyZeroDec()
	}

	num := sdkmath.LegacyNewDecFromInt(
		sdkmath.NewIntFromUint64(gasUsed).Sub(sdkmath.NewIntFromUint64(gasTarget)).Abs(),
	)
	num = num.Mul(baseFee)
	num = num.QuoInt(sdkmath.NewIntFromUint64(gasTarget))
	num = num.QuoInt(sdkmath.NewIntFromUint64(baseFeeChangeDenom))

	if gasUsed > gasTarget {
		delta := sdkmath.LegacyMaxDec(num, minUnitGas)
		return baseFee.Add(delta)
	}
	if minGasPrice.IsNil() {
		minGasPrice = sdkmath.LegacyZeroDec()
	}
	return sdkmath.LegacyMaxDec(baseFee.Sub(num), minGasPrice)
}
