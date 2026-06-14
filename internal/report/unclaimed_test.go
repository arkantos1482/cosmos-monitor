package report

import (
	"testing"

	"github.com/arkantos1482/cosmos-monitor/internal/fetch"
)

func TestDetectOutstandingModelTotalEscrow(t *testing.T) {
	model := detectOutstandingModel(0.006854, 0.006854, 0.000685)
	if model != outstandingIsTotalEscrow {
		t.Fatalf("expected total-escrow model, got %v", model)
	}
}

func TestDetectOutstandingModelDelegatorShare(t *testing.T) {
	model := detectOutstandingModel(0.007539, 0.006854, 0.000685)
	if model != outstandingIsDelegatorShare {
		t.Fatalf("expected delegator-share model, got %v", model)
	}
}

func TestComputeUnclaimedTotalsTotalEscrow(t *testing.T) {
	vals := []fetch.ValidatorInfo{{
		OutstandingRewardsAmt: "6854164346049315",
		OutstandingRewardsDenom: "apmt",
		CommissionEarnedAmt:   "685416434604931",
		CommissionEarnedDenom: "apmt",
	}}
	mods := []fetch.ModuleBalanceInfo{{
		Name: "distribution", Amount: "6854164346049315", Denom: "apmt",
	}}
	tot := computeUnclaimedTotals(vals, mods)
	if tot.Model != outstandingIsTotalEscrow {
		t.Fatalf("model=%v", tot.Model)
	}
	if tot.TotalF <= 0 || tot.DelegatorF <= 0 || tot.CommissionF <= 0 {
		t.Fatalf("expected positive totals, got %+v", tot)
	}
	if !amountsClose(tot.TotalF, tot.DelegatorF+tot.CommissionF) {
		t.Fatalf("total should equal delegator+commission: %+v", tot)
	}
}
