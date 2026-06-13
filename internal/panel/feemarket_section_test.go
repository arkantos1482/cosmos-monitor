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
	}
	out := BuildView(ViewFeemarket, d)
	for _, want := range []string{
		`dash-section--feemarket`,
		`class="fm-summary"`,
		`fm-mechanics`,
		`eco-domain--feemarket`,
		`projected next base fee`,
		`Live state`,
		`EIP-1559 mechanics`,
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
}
