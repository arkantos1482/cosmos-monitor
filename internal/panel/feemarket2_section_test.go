package panel

import (
	"strings"
	"testing"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func TestWriteFeemarket2Section(t *testing.T) {
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
	out := BuildView(ViewFeemarket2, d)
	for _, want := range []string{
		`dash-section--feemarket2`,
		`class="fm2-summary"`,
		`fm2-mechanics`,
		`eco-domain--feemarket2`,
		`projected next base fee`,
		`Live state`,
		`EIP-1559 mechanics`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in output", want)
		}
	}
}

func TestFeemarket2NavRegistered(t *testing.T) {
	found := false
	for _, item := range Nav {
		if item.View == ViewFeemarket2 {
			found = true
			if item.Path != "/s/feemarket2" {
				t.Fatalf("path=%s", item.Path)
			}
		}
	}
	if !found {
		t.Fatal("feemarket2 not in nav")
	}
}
