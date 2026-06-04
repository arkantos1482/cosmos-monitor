package panel

import (
	"strings"
	"testing"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func TestWriteFeemarketSectionHTML(t *testing.T) {
	var b strings.Builder
	w := newWriter(&b)
	d := model.Report{
		BlockHeight: "100", BaseFee: "1 PMT", BaseFeeRaw: "1000000000000",
		BlockGas: "21000", BlockGasLimit: 100_000_000, Elasticity: 2,
		BaseFeeChangeDenominator: 8,
		ParentBlockGasUsed: 21000, ParentBlockGasWanted: 21000,
		ParentBlockResultsOK: true, GasPrice: "1000000000000", EVMDenom: "apmt",
	}
	writeFeemarketSection(w, d)
	w.flush()
	out := b.String()
	for _, want := range []string{
		`class="fee-traffic"`,
		`class="fee-badge`,
		`class="fee-meter"`,
		`class="fee-hero-line"`,
		`class="fee-cards"`,
		`Last block (N−1)`,
		`Formula`,
		`This block (N)`,
		`Params`,
		`FEE FALLING`,
		`gas_wanted`,
		`block_results`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in output", want)
		}
	}
	for _, bad := range []string{
		`math-panel`, `mermaid`, `data-tex-b64`,
		`Receipt`, `Adjustment logic`, `fee-traffic-stats`,
		`fee-key-metrics`, `class="code-block"`,
		`Chain parameters`,
	} {
		if strings.Contains(out, bad) {
			t.Fatalf("fee section should not contain %q", bad)
		}
	}
	if strings.Count(out, `class="stat-grid"`) != 0 {
		t.Fatal("fee section should not use stat-grid key metrics")
	}
}

func TestWriteFeeMathHTML(t *testing.T) {
	var b strings.Builder
	w := newWriter(&b)
	w.MathLatex(`\[ W_{\text{stored}} = 1 \]`)
	out := b.String()
	if !strings.Contains(out, `math-line`) || !strings.Contains(out, `math-panel`) {
		t.Fatalf("expected math-panel/math-line, got %q", out)
	}
}
