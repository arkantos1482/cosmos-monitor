package feemarket

import (
	"strings"
	"testing"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func TestLoadStateEIP1559Rising(t *testing.T) {
	s := LoadState(model.Report{
		BlockHeight:              "100",
		BaseFee:                  "1.125 apmt",
		BaseFeeRaw:               "1125000000",
		BlockGas:                 "60000000",
		BlockGasLimit:            100_000_000,
		Elasticity:               2,
		BaseFeeChangeDenominator: 8,
		EVMDenom:                 "apmt",
	})
	if !s.EIP1559On {
		t.Fatal("expected EIP-1559 on")
	}
	if s.Adj != AdjRising {
		t.Fatalf("adj=%v want rising", s.Adj)
	}
	if s.GasTarget != 50_000_000 {
		t.Fatalf("target=%d", s.GasTarget)
	}
}

func TestLoadStateNoBaseFee(t *testing.T) {
	s := LoadState(model.Report{NoBaseFee: true, BlockHeight: "10"})
	if s.Adj != AdjDisabled {
		t.Fatalf("adj=%v", s.Adj)
	}
}

func TestLoadStateUnlimitedMaxGasNoPanic(t *testing.T) {
	s := LoadState(model.Report{
		BlockHeight:              "100",
		BaseFeeRaw:               "1125000000",
		BlockGasLimit:            ^uint64(0),
		Elasticity:               2,
		BaseFeeChangeDenominator: 8,
		EVMDenom:                 "apmt",
	})
	if !s.UnlimitedBlockGas {
		t.Fatal("expected unlimited block gas")
	}
	if s.HasProjection {
		t.Fatal("unlimited max_gas should not project base fee")
	}
}

func TestTransferCost(t *testing.T) {
	got := TransferCost("1000", "apmt")
	if got == "" || got == "—" {
		t.Fatalf("transfer cost: %q", got)
	}
}

func TestTransferCostTinyBaseFee(t *testing.T) {
	// 7 apmt base fee → 7e-18 PMT; 21k gas → 1.47e-13 PMT
	got := TransferCost("0.000000000000000007", "apmt")
	if got == "" || got == "—" || got == "0" || got == "0 PMT" {
		t.Fatalf("transfer cost: %q", got)
	}
	if !strings.Contains(got, "1.47e-13") {
		t.Fatalf("transfer cost = %q, want 1.47e-13 PMT", got)
	}
}
