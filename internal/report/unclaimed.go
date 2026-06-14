package report

import (
	"math"

	"github.com/arkantos1482/cosmos-monitor/internal/fetch"
)

// outstandingRewardsModel describes how Σ outstanding_rewards relates to the
// distribution module bank balance on this chain.
type outstandingRewardsModel int

const (
	outstandingUnknown outstandingRewardsModel = iota
	// outstandingIsTotalEscrow: bank ≈ Σ outstanding_rewards. On PMT/precisebank
	// the REST outstanding amount is full per-validator escrow; commission is a
	// subset, so delegator share = outstanding − commission per validator.
	outstandingIsTotalEscrow
	// outstandingIsDelegatorShare: bank ≈ Σ outstanding + Σ commission (standard cosmos-sdk).
	outstandingIsDelegatorShare
)

func amountsClose(a, b float64) bool {
	if a == 0 && b == 0 {
		return true
	}
	diff := math.Abs(a - b)
	if diff < 1e-12 {
		return true
	}
	scale := math.Max(math.Abs(a), math.Abs(b))
	return scale > 0 && diff/scale < 1e-6
}

func distributionBankFloat(mods []fetch.ModuleBalanceInfo) float64 {
	for _, mod := range mods {
		if mod.Name == "distribution" && mod.Amount != "" {
			f, _ := fetch.NormalizeCoin(mod.Amount, mod.Denom)
			return f
		}
	}
	return 0
}

func detectOutstandingModel(distBankF, totalOutF, totalCommF float64) outstandingRewardsModel {
	if totalOutF <= 0 && totalCommF <= 0 {
		return outstandingUnknown
	}
	if distBankF > 0 {
		if amountsClose(distBankF, totalOutF) {
			return outstandingIsTotalEscrow
		}
		if amountsClose(distBankF, totalOutF+totalCommF) {
			return outstandingIsDelegatorShare
		}
	}
	return outstandingIsDelegatorShare
}

func sumOutstandingCommission(vals []fetch.ValidatorInfo) (totalOutF, totalCommF float64, denom string) {
	for _, v := range vals {
		if v.OutstandingRewardsAmt != "" {
			f, dd := fetch.NormalizeCoin(v.OutstandingRewardsAmt, v.OutstandingRewardsDenom)
			totalOutF += f
			denom = dd
		}
		if v.CommissionEarnedAmt != "" {
			f, dd := fetch.NormalizeCoin(v.CommissionEarnedAmt, v.CommissionEarnedDenom)
			totalCommF += f
			if denom == "" {
				denom = dd
			}
		}
	}
	return totalOutF, totalCommF, denom
}

type unclaimedTotals struct {
	DelegatorF  float64
	CommissionF float64
	TotalF      float64
	Denom       string
	Model       outstandingRewardsModel
}

func computeUnclaimedTotals(vals []fetch.ValidatorInfo, mods []fetch.ModuleBalanceInfo) unclaimedTotals {
	totalOutF, totalCommF, denom := sumOutstandingCommission(vals)
	model := detectOutstandingModel(distributionBankFloat(mods), totalOutF, totalCommF)

	var delegatorF, totalF float64
	switch model {
	case outstandingIsTotalEscrow:
		for _, v := range vals {
			outF, _ := fetch.NormalizeCoin(v.OutstandingRewardsAmt, v.OutstandingRewardsDenom)
			commF, _ := fetch.NormalizeCoin(v.CommissionEarnedAmt, v.CommissionEarnedDenom)
			if d := outF - commF; d > 0 {
				delegatorF += d
			}
		}
		totalF = totalOutF
	default:
		delegatorF = totalOutF
		totalF = totalOutF + totalCommF
	}

	return unclaimedTotals{
		DelegatorF:  delegatorF,
		CommissionF: totalCommF,
		TotalF:      totalF,
		Denom:       denom,
		Model:       model,
	}
}

func validatorDelegatorOutstanding(v fetch.ValidatorInfo, model outstandingRewardsModel) string {
	if v.OutstandingRewardsAmt == "" {
		return v.OutstandingRewards
	}
	outF, denom := fetch.NormalizeCoin(v.OutstandingRewardsAmt, v.OutstandingRewardsDenom)
	if model != outstandingIsTotalEscrow {
		return v.OutstandingRewards
	}
	commF, _ := fetch.NormalizeCoin(v.CommissionEarnedAmt, v.CommissionEarnedDenom)
	delF := outF - commF
	if delF < 0 {
		delF = 0
	}
	return fetch.FormatAmountUnit(delF, denom)
}
