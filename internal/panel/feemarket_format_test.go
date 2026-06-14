package panel

import (
	"strings"
	"testing"

	"github.com/arkantos1482/cosmos-monitor/internal/feemarket"
	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func TestFmGasLimitHTMLUnlimited(t *testing.T) {
	s := feemarket.LoadState(model.Report{BlockGasLimit: ^uint64(0)})
	out := fmGasLimitHTML(s)
	for _, want := range []string{`(-1) ∞`, `fm-sentinel__raw`, `18446744073709551615`} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in %q", want, out)
		}
	}
}

func TestFmGasLimitHTMLFinite(t *testing.T) {
	s := feemarket.LoadState(model.Report{BlockGasLimit: 30_000_000})
	out := fmGasLimitHTML(s)
	if out != "30000000 gas" {
		t.Fatalf("got %q", out)
	}
	if strings.Contains(out, "fm-sentinel") {
		t.Fatal("finite limit should not use sentinel styling")
	}
}

func TestFmGasTargetHTMLUnlimited(t *testing.T) {
	s := feemarket.LoadState(model.Report{
		BlockGasLimit: ^uint64(0),
		Elasticity:    2,
	})
	out := fmGasTargetHTML(s)
	for _, want := range []string{`max ÷ 2`, `fm-sentinel__raw`, `9223372036854775807`} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in %q", want, out)
		}
	}
}

func TestFmGasTargetHTMLFinite(t *testing.T) {
	s := feemarket.LoadState(model.Report{
		BlockGasLimit: 100_000_000,
		Elasticity:    2,
	})
	out := fmGasTargetHTML(s)
	if out != "50000000 gas" {
		t.Fatalf("got %q", out)
	}
}

func TestFmMechanicsVarsHTML(t *testing.T) {
	s := feemarket.LoadState(model.Report{
		BlockHeight:              "100",
		BaseFee:                  "1.125 apmt",
		BaseFeeRaw:               "1125000000",
		BlockGasLimit:            100_000_000,
		Elasticity:               2,
		BaseFeeChangeDenominator: 8,
		MinGasMultiplier:         "0.5",
		MinGasPrice:              "0 apmt",
		EVMDenom:                 "apmt",
		ParentBlockGasUsed:       55_000_000,
		ParentBlockTxGasWanted:   120_000_000,
		ParentBlockGasWanted:     60_000_000,
		ParentBlockResultsOK:     true,
	})
	out := fmMechanicsVarsHTML(s)
	for _, want := range []string{
		`gas_used`,
		`gas_wanted`,
		`min_gas_multiplier`,
		`>W<`,
		`block_gas_limit`,
		`elasticity_multiplier`,
		`target`,
		`|W−target|`,
		`10000000 gas`,
		`base_fee`,
		`denom`,
		`min_fee_step`,
		`min_gas_price`,
		`55000000 gas`,
		`120000000 gas`,
		`60000000 gas`,
		`50000000 gas`,
		`1.125 apmt`,
		`>2<`,
		`>8<`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in %q", want, out)
		}
	}
	// W = max(55M, 120M × 0.5) = max(55M, 60M) = 60M; |W−target| = 10M
	if s.GasWanted != 60_000_000 {
		t.Fatalf("GasWanted=%d want 60M", s.GasWanted)
	}
}
