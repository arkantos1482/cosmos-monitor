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
		`class="fee-cards"`,
		`class="code-block"`,
		`Chain parameters`,
		`Receipt walkthrough`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in output", want)
		}
	}
	for _, bad := range []string{`math-panel`, `mermaid`, `data-tex-b64`} {
		if strings.Contains(out, bad) {
			t.Fatalf("fee section should not contain %q", bad)
		}
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
