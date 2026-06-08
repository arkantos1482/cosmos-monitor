package panel

import (
	"strings"
	"testing"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func feemarketChunk(t *testing.T, out string) string {
	t.Helper()
	idx := strings.Index(out, "6. FEE MARKET")
	end := strings.Index(out, "7. GOVERNANCE")
	if idx < 0 || end < 0 {
		t.Fatal("expected fee market and governance sections")
	}
	return out[idx:end]
}

func TestWriteFeemarketStoryLayout(t *testing.T) {
	d := model.Report{
		BlockHeight: "100", BaseFee: "1", BaseFeeRaw: "1000",
		BlockGas: "21000", ParentBlockGasWanted: 21000, ParentBlockGasUsed: 18000,
		BlockGasLimit: 100_000_000, Elasticity: 2,
		BaseFeeChangeDenominator: 8, MinGasMultiplier: "0.5", MinGasPrice: "0.01",
		ParentBlockResultsOK: true, GasPrice: "1000",
	}
	chunk := feemarketChunk(t, Build(d))
	for _, want := range []string{
		`class="fee-hero"`,
		`class="fee-pipeline"`,
		`class="fee-cards"`,
		`fee-card--demand`,
		`fee-card--adjust`,
		`fee-card--network`,
		`fee-card--wallet`,
		"Block demand (block N−1)",
		"Fee adjustment",
		"Network fee now (block N)",
		"Wallet quote",
		"Demand vs capacity",
		`id="feemarket-ref"`,
		`class="dash-details"`,
		"Parameters, formulas &amp; data sources",
		"Block N−1",
		"BeginBlock",
	} {
		if !strings.Contains(chunk, want) {
			t.Fatalf("fee market chunk missing %q", want)
		}
	}
	refIdx := strings.Index(chunk, `id="feemarket-ref"`)
	if refIdx < 0 {
		t.Fatal("missing feemarket reference details")
	}
	aboveRef := chunk[:refIdx]
	for _, gone := range []string{
		`fee-key-metrics`,
		`class="dash-subheading">Variables</h3>`,
		`class="dash-subheading">Params</h3>`,
		"W / target",
	} {
		if strings.Contains(aboveRef, gone) {
			t.Fatalf("fee market body above reference should not contain %q", gone)
		}
	}
	if !strings.Contains(chunk, "Symbols") {
		t.Fatal("reference block should include symbol glossary")
	}
}

func TestWriteFeemarketNoBaseFee(t *testing.T) {
	d := model.Report{
		BlockHeight: "50", BaseFee: "0.5", NoBaseFee: true,
		GasPrice: "500",
	}
	chunk := feemarketChunk(t, Build(d))
	if strings.Count(chunk, `class="fee-pipeline__step"`) != 2 {
		t.Fatal("no_base_fee pipeline should have 2 steps")
	}
	if strings.Count(chunk, `class="fee-card `) != 2 {
		t.Fatal("no_base_fee should render 2 story cards")
	}
	if !strings.Contains(chunk, "FIXED PRICING") {
		t.Fatal("no_base_fee should show FIXED PRICING badge")
	}
}

func TestWriteFeemarketUnlimitedMaxGas(t *testing.T) {
	d := model.Report{
		BlockHeight: "200", BaseFee: "2", BaseFeeRaw: "2000",
		BlockGas: "21000", ParentBlockGasWanted: 21000, ParentBlockGasUsed: 21000,
		BlockGasLimit: ^uint64(0), Elasticity: 2,
		BaseFeeChangeDenominator: 8, MinGasMultiplier: "1",
		ParentBlockResultsOK: true,
	}
	chunk := feemarketChunk(t, Build(d))
	if strings.Contains(chunk, `aria-label="Demand vs capacity"`) {
		t.Fatal("unlimited max_gas should hide utilization meter")
	}
	if !strings.Contains(chunk, "MaxUint64") {
		t.Fatal("unlimited max_gas should explain MaxUint64 sentinel")
	}
	if !strings.Contains(chunk, `class="fee-pipeline"`) {
		t.Fatal("unlimited max_gas should still render pipeline")
	}
}
