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
		`class="fee-hero"`,
		`class="fee-traffic"`,
		`class="fee-badge`,
		`class="fee-meter"`,
		`class="fee-hero-line"`,
		`fee-key-metrics`,
		`fee-kpi`,
		`class="fee-formula"`,
		`Variables`,
		`Formulas`,
		`Params`,
		`FEE FALLING`,
		`Stored gas wanted`,
		`Live value`,
		`Utilization`,
		`Next adjustment`,
		`Base fee`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in output", want)
		}
	}
	for _, bad := range []string{
		`math-panel`, `mermaid`, `data-tex-b64`,
		`Receipt`, `Adjustment logic`, `fee-traffic-stats`,
		`Chain parameters`,
		`Last block (N−1)`,
		`How fees adjust`,
		`This block (N)`,
		`What wallets see`,
	} {
		if strings.Contains(out, bad) {
			t.Fatalf("fee section should not contain %q", bad)
		}
	}
	if strings.Count(out, `class="kpi-grid"`) != 0 {
		t.Fatal("fee section should not use kpi-grid for key metrics")
	}
}

func TestWriteFeemarketSectionUnlimitedMaxGas(t *testing.T) {
	var b strings.Builder
	w := newWriter(&b)
	d := model.Report{
		BlockHeight: "100", BaseFee: "1 PMT", BaseFeeRaw: "1000000000000",
		BlockGas: "21000", BlockGasLimit: ^uint64(0), Elasticity: 2,
		BaseFeeChangeDenominator: 8,
		ParentBlockGasUsed: 21000, ParentBlockGasWanted: 21000,
		ParentBlockResultsOK: true, EVMDenom: "apmt",
	}
	writeFeemarketSection(w, d)
	w.flush()
	out := b.String()
	mystery := "9,223,372,036,854,775,807"
	if strings.Contains(out, mystery) {
		t.Fatalf("HTML should not show raw MaxUint64 decimal: %s", out)
	}
	if strings.Contains(out, `class="fee-meter"`) {
		t.Fatal("unlimited max_gas should hide load meter")
	}
	if !strings.Contains(out, "MaxUint64") || !strings.Contains(out, "sentinel") {
		t.Fatal("unlimited section should label MaxUint64 sentinel")
	}
}
