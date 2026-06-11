package panel

import (
	"strings"
	"testing"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func feemarketChunk(t *testing.T, out string) string {
	t.Helper()
	idx := strings.Index(out, `class="dash-heading">3. FEE MARKET</h2>`)
	end := strings.Index(out, "4. GOVERNANCE")
	if idx < 0 || end < 0 {
		t.Fatal("expected fee market and governance sections")
	}
	return out[idx:end]
}

func TestWriteFeemarketLadderLayout(t *testing.T) {
	d := model.Report{
		BlockHeight: "100", BaseFee: "1", BaseFeeRaw: "1000",
		BlockGas: "21000", ParentBlockGasWanted: 21000, ParentBlockGasUsed: 18000,
		BlockGasLimit: 100_000_000, Elasticity: 2,
		BaseFeeChangeDenominator: 8, MinGasMultiplier: "0.5", MinGasPrice: "0.01",
		ParentBlockResultsOK: true, EVMDenom: "apmt",
	}
	chunk := feemarketChunk(t, Build(d))

	for _, want := range []string{
		`id="fee-L1"`,
		`id="fee-L2"`,
		`id="fee-L3"`,
		`id="fee-L4"`,
		`id="fee-L5"`,
		`class="fee-level"`,
		`class="fee-nav"`,
		`class="fee-summary"`,
		"What you pay now",
		"Why the fee moved",
		"What the chain measured",
		"When each value is written",
		"Formula &amp; parameters",
		"Illustrative example: when W ≠ gas_used",
		"Three pools",
		"In-block accumulator",
	} {
		if !strings.Contains(chunk, want) {
			t.Fatalf("fee market chunk missing %q", want)
		}
	}

	for _, gone := range []string{
		`class="fee-flow"`,
		`class="fee-hero"`,
		`id="feemarket-ref"`,
		"mempool ×",
		"Finance",
		"Operator",
		"Developer",
		"6. FEE MARKET",
	} {
		if strings.Contains(chunk, gone) {
			t.Fatalf("fee market should not contain %q", gone)
		}
	}

	l1Start := strings.Index(chunk, `id="fee-L1"`)
	l1End := strings.Index(chunk, `id="fee-L2"`)
	if l1Start < 0 || l1End < 0 {
		t.Fatal("missing L1 or L2")
	}
	l1 := chunk[l1Start:l1End]
	for _, forbidden := range []string{">W<", "gas_used", "target"} {
		if strings.Contains(strings.ToLower(l1), forbidden) {
			t.Fatalf("L1 should not expose %q", forbidden)
		}
	}
}

func TestWriteFeemarketAtFloorBadge(t *testing.T) {
	d := model.Report{
		BlockHeight: "1284501", BaseFee: "7 apmt", BaseFeeRaw: "7",
		BlockGasLimit: ^uint64(0), Elasticity: 2,
		BaseFeeChangeDenominator: 8, MinGasMultiplier: "0.5",
		ParentBlockGasWanted: 2_847_392, ParentBlockGasUsed: 2_847_392,
		MinGasPriceRaw: "0", EVMDenom: "apmt",
	}
	chunk := feemarketChunk(t, Build(d))
	if !strings.Contains(chunk, "AT FLOOR") {
		t.Fatal("expected AT FLOOR badge")
	}
	if strings.Contains(chunk, `fee-badge fee-badge--falling`) {
		t.Fatal("AT FLOOR must not use falling badge class")
	}
	if strings.Contains(chunk, "decrease step base÷8 = 0 apmt") {
		t.Fatal("AT FLOOR footnote should explain sub-integer truncate, not bare 0 apmt")
	}
	if !strings.Contains(chunk, "truncates to 0") {
		t.Fatal("AT FLOOR footnote should mention truncate-to-zero precision")
	}
}

func TestWriteFeemarketNoBaseFee(t *testing.T) {
	d := model.Report{
		BlockHeight: "50", BaseFee: "0.5", NoBaseFee: true,
	}
	chunk := feemarketChunk(t, Build(d))
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
	if !strings.Contains(chunk, "MaxUint64") {
		t.Fatal("unlimited max_gas should explain MaxUint64 sentinel")
	}
	if strings.Contains(chunk, `aria-label="Demand vs capacity"`) {
		t.Fatal("unlimited max_gas should not show legacy demand meter label")
	}
}

func TestWriteFeemarketFiniteMaxGasMeter(t *testing.T) {
	d := model.Report{
		BlockHeight: "100", BaseFeeRaw: "1000",
		BlockGasLimit: 30_000_000, Elasticity: 2,
		BaseFeeChangeDenominator: 8,
		ParentBlockGasWanted: 12_400_000,
	}
	chunk := feemarketChunk(t, Build(d))
	if !strings.Contains(chunk, `aria-label="Demand vs target"`) {
		t.Fatal("finite max_gas should show demand meter")
	}
}

func TestWriteFeemarketEnableHeightBanner(t *testing.T) {
	d := model.Report{
		BlockHeight: "50", EnableHeight: 100,
		BlockGasLimit: 30_000_000, Elasticity: 2,
	}
	chunk := feemarketChunk(t, Build(d))
	if !strings.Contains(chunk, "FEES DISABLED") {
		t.Fatal("expected FEES DISABLED badge")
	}
}

func TestWriteFeemarketL5ChainParams(t *testing.T) {
	d := model.Report{
		BlockHeight: "1", BaseFeeRaw: "7",
		BlockGasLimit: ^uint64(0), Elasticity: 2,
		BaseFeeChangeDenominator: 8,
		MaxBlockBytes: 22_020_096,
	}
	chunk := feemarketChunk(t, BuildWithOptions(d, Options{ShowSources: true}))
	for _, want := range []string{
		"22,020,096",
		"min_unit_gas",
		`class="dash-sources"`,
		`>Data sources</summary>`,
		`class="dash-callout dash-callout--hint hint"`,
		`class="hint-provenance"`,
		`CometBFT GET /block_results`,
		"consensus_params",
		"vm/v1/config",
		"§ Validator",
	} {
		if !strings.Contains(chunk, want) {
			t.Fatalf("L5 missing %q", want)
		}
	}
	for _, gone := range []string{
		"minimum-gas-prices",
		"Node acceptance",
	} {
		if strings.Contains(chunk, gone) {
			t.Fatalf("fee market L5 should not contain node setting %q", gone)
		}
	}
	if strings.Contains(chunk, `fee-level__note">gas_used`) {
		t.Fatal("fee market data sources should use section provenance, not fee-level__note paragraph")
	}
	l5Start := strings.Index(chunk, `id="fee-L5"`)
	if l5Start < 0 {
		t.Fatal("missing L5")
	}
	l5CloseRel := strings.Index(chunk[l5Start:], `</details>`)
	if l5CloseRel < 0 {
		t.Fatal("missing L5 closing details")
	}
	l5Block := chunk[l5Start : l5Start+l5CloseRel]
	if strings.Contains(l5Block, "CometBFT GET /block_results") {
		t.Fatal("L5 should not embed data source provenance")
	}
	sourcesIdx := strings.Index(chunk, `class="dash-sources"`)
	if sourcesIdx < 0 || sourcesIdx < l5Start+l5CloseRel {
		t.Fatal("data sources should render in section footer after L5")
	}
}

func TestWriteFeemarketSourcesHiddenByDefault(t *testing.T) {
	d := model.Report{
		BlockHeight: "1", BaseFeeRaw: "7",
		BlockGasLimit: ^uint64(0), Elasticity: 2,
		BaseFeeChangeDenominator: 8, MinGasMultiplier: "0.5",
		ParentBlockResultsOK: true, EVMDenom: "apmt",
	}
	chunk := feemarketChunk(t, Build(d))
	if strings.Contains(chunk, `class="dash-sources"`) {
		t.Fatal("fee market should omit data sources without -show-sources")
	}
}

func TestWriteNodeFeeAcceptance(t *testing.T) {
	d := model.Report{
		NodeMinGasPrices: "0apmt", NodeEVMMinTip: "0",
		NodeMempoolPriceLimit: "1", NodeMaxTxGasWanted: "0",
		NodeAppTomlPath: "/home/ubuntu/.evmd/config/app.toml",
		Local: model.LocalValidator{IsValidator: true, Moniker: "node1", Status: "BOND_STATUS_BONDED"},
	}
	out := BuildView(ViewNode, d)
	for _, want := range []string{
		"Fee acceptance (app.toml)",
		"minimum-gas-prices",
		"evm.min-tip",
		"evm.mempool.price-limit",
		"evm.max-tx-gas-wanted",
		"/home/ubuntu/.evmd/config/app.toml",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("validator section missing %q", want)
		}
	}
	infra := BuildView(ViewInfra, d)
	if strings.Contains(infra, "Fee acceptance (app.toml)") {
		t.Fatal("infra section should not contain app.toml fee acceptance")
	}
}
