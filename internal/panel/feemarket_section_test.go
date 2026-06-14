package panel

import (
	"strings"
	"testing"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func TestWriteFeemarketSection(t *testing.T) {
	d := model.Report{
		BlockHeight:              "100",
		BaseFee:                  "1.125 apmt",
		BaseFeeRaw:               "1125000000",
		BlockGas:                 "60000000",
		BlockGasLimit:            100_000_000,
		Elasticity:               2,
		BaseFeeChangeDenominator: 8,
		MinGasMultiplier:         "0.5",
		EVMDenom:                 "apmt",
		NodeMinGasPrices:         "0.025apmt",
		ParentBlockGasUsed:       55_000_000,
		ParentBlockTxGasWanted:   120_000_000,
		ParentBlockGasWanted:     60_000_000,
		ParentBlockResultsOK:     true,
	}
	out := BuildView(ViewFeemarket, d)
	for _, want := range []string{
		`dash-section--feemarket`,
		`class="fm-summary"`,
		`fm-mechanics`,
		`fm-mechanics__vars`,
		`eco-domain--feemarket`,
		`projected next base fee`,
		`gas used`,
		`gas wanted`,
		`floor for W`,
		`min_gas_multiplier for W`,
		`demand vs target`,
		`Live state`,
		`EIP-1559 mechanics`,
		`formula input`,
		`55000000 gas`,
		`60000000 gas`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in output", want)
		}
	}
}

func TestFeemarketNavRegistered(t *testing.T) {
	found := false
	for _, item := range Nav {
		if item.View == ViewFeemarket {
			found = true
			if item.Path != "/s/feemarket" {
				t.Fatalf("path=%s", item.Path)
			}
		}
	}
	if !found {
		t.Fatal("feemarket not in nav")
	}
}

func TestWriteFeemarketUnlimitedMaxGas(t *testing.T) {
	d := model.Report{
		BlockHeight:              "200",
		BaseFee:                  "2 apmt",
		BaseFeeRaw:               "2000",
		BlockGasLimit:            ^uint64(0),
		Elasticity:               2,
		BaseFeeChangeDenominator: 8,
		EVMDenom:                 "apmt",
	}
	out := BuildView(ViewFeemarket, d)
	if !strings.Contains(out, `dash-section--feemarket`) {
		t.Fatal("missing feemarket section")
	}
	if strings.Contains(out, `projected next base fee`) {
		t.Fatal("unlimited max_gas should not show projected base fee")
	}
	for _, want := range []string{`(-1) ∞`, `max ÷ 2`, `fm-sentinel__raw`} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in unlimited max_gas view", want)
		}
	}
	liveIdx := strings.Index(out, "Live state")
	mechIdx := strings.Index(out, "EIP-1559 mechanics")
	if liveIdx < 0 || mechIdx < 0 || mechIdx <= liveIdx {
		t.Fatal("missing live state or mechanics subsection")
	}
	liveSlice := out[liveIdx:mechIdx]
	for _, bad := range []string{"block gas limit"} {
		if strings.Contains(liveSlice, bad) {
			t.Fatalf("static %q should not appear in live state", bad)
		}
	}
}
