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

func feemarketAboveRef(t *testing.T, chunk string) string {
	t.Helper()
	refIdx := strings.Index(chunk, `id="feemarket-ref"`)
	if refIdx < 0 {
		t.Fatal("missing feemarket reference details")
	}
	return chunk[:refIdx]
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
	aboveRef := feemarketAboveRef(t, chunk)

	for _, want := range []string{
		`class="fee-hero"`,
		`class="fee-flow"`,
		`fee-flow__step--demand`,
		`fee-flow__step--adjust`,
		`fee-flow__step--network`,
		`fee-flow__step--wallet`,
		"Block N−1",
		"vs target",
		"BeginBlock",
		"Wallet RPC",
		"Demand vs capacity",
		"Below target",
		"|Δbase|",
		`id="feemarket-ref"`,
		`class="dash-details"`,
		"Parameters, formulas &amp; data sources",
	} {
		if !strings.Contains(chunk, want) {
			t.Fatalf("fee market chunk missing %q", want)
		}
	}
	for _, gone := range []string{
		`class="fee-pipeline"`,
		`class="fee-cards"`,
		`fee-key-metrics`,
		`class="dash-subheading">Variables</h3>`,
		`class="dash-subheading">Params</h3>`,
		"W / target",
	} {
		if strings.Contains(aboveRef, gone) {
			t.Fatalf("fee market body above reference should not contain %q", gone)
		}
	}
	if strings.Count(aboveRef, `class="fee-flow__step `) != 4 {
		t.Fatalf("expected 4 flow steps above reference, got %d", strings.Count(aboveRef, `class="fee-flow__step `))
	}
	if strings.Count(aboveRef, ">gas_used<") != 1 {
		t.Fatalf("gas_used should appear once above reference, got %d", strings.Count(aboveRef, ">gas_used<"))
	}
	ex := buildFeemarketExplain(d)
	if ex.SummaryLine != "" && strings.Contains(aboveRef, ex.SummaryLine) {
		// SummaryLine is in hero only; flow must not repeat the narrative sentence.
		flowStart := strings.Index(aboveRef, `class="fee-flow"`)
		if flowStart >= 0 && strings.Contains(aboveRef[flowStart:], ex.SummaryLine) {
			t.Fatal("flow should not duplicate hero SummaryLine")
		}
	}
	if !strings.Contains(chunk, "Symbols") {
		t.Fatal("reference block should include symbol glossary")
	}
	if !strings.Contains(chunk, "Chain parameters") {
		t.Fatal("reference block should include chain parameters")
	}
}

func TestWriteFeemarketNoBaseFee(t *testing.T) {
	d := model.Report{
		BlockHeight: "50", BaseFee: "0.5", NoBaseFee: true,
		GasPrice: "500",
	}
	chunk := feemarketChunk(t, Build(d))
	aboveRef := feemarketAboveRef(t, chunk)
	if strings.Count(aboveRef, `class="fee-flow__step `) != 2 {
		t.Fatalf("no_base_fee flow should have 2 steps, got %d", strings.Count(aboveRef, `class="fee-flow__step `))
	}
	if !strings.Contains(chunk, "FIXED PRICING") {
		t.Fatal("no_base_fee should show FIXED PRICING badge")
	}
	if strings.Contains(aboveRef, `class="fee-pipeline"`) || strings.Contains(aboveRef, `class="fee-cards"`) {
		t.Fatal("no_base_fee should not render legacy pipeline/cards")
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
	aboveRef := feemarketAboveRef(t, chunk)
	if strings.Contains(aboveRef, `aria-label="Demand vs capacity"`) {
		t.Fatal("unlimited max_gas should hide utilization meter")
	}
	if !strings.Contains(aboveRef, "MaxUint64") {
		t.Fatal("unlimited max_gas should explain MaxUint64 sentinel")
	}
	if !strings.Contains(aboveRef, `class="fee-flow"`) {
		t.Fatal("unlimited max_gas should still render merged flow")
	}
}
