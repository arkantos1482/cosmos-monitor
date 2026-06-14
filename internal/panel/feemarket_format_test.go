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
