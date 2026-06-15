package report

import (
	"testing"

	"github.com/arkantos1482/cosmos-monitor/internal/fetch"
	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func buildBondedReport(chain fetch.ChainSnapshot) model.Report {
	return Build(chain, fetch.EVMSnapshot{}, fetch.SystemSnapshot{}, fetch.DockerSnapshot{}, "", model.StatusAvailability{}, fetch.AppTomlGasConfig{}, nil)
}

func TestBondedPctFromTotalSupply(t *testing.T) {
	chain := fetch.ChainSnapshot{
		Params:           fetch.ChainParams{BondDenom: "apmt"},
		BondedTokens:     "400000000000000000000000000",
		NotBondedTokens:  "0",
		TotalSupply:      "2000000000000000000000000000",
		TotalSupplyDenom: "apmt",
	}
	d := buildBondedReport(chain)

	if d.BondedPct != 20 {
		t.Fatalf("BondedPct = %v, want 20 (bonded ÷ total supply)", d.BondedPct)
	}
	if d.MintBondedPct != 100 {
		t.Fatalf("MintBondedPct = %v, want 100 (pool ratio when not_bonded empty)", d.MintBondedPct)
	}
}

func TestBondedPctFromStakingPoolWhenNoSupply(t *testing.T) {
	chain := fetch.ChainSnapshot{
		Params:          fetch.ChainParams{BondDenom: "apmt"},
		BondedTokens:    "4000000000000000000000000",
		NotBondedTokens: "8000000000000000000000000",
	}
	d := buildBondedReport(chain)

	const want = 33.333333333333336
	if d.BondedPct < want-0.01 || d.BondedPct > want+0.01 {
		t.Fatalf("BondedPct = %v, want ~33.33 from pool when supply unavailable", d.BondedPct)
	}
	if d.MintBondedPct < want-0.01 || d.MintBondedPct > want+0.01 {
		t.Fatalf("MintBondedPct = %v, want ~33.33", d.MintBondedPct)
	}
}

func TestBondedPctPrefersTotalSupplyOverPool(t *testing.T) {
	chain := fetch.ChainSnapshot{
		Params:           fetch.ChainParams{BondDenom: "apmt"},
		BondedTokens:     "1000000000000000000000000",
		NotBondedTokens:  "1000000000000000000000000",
		TotalSupply:      "4000000000000000000000000",
		TotalSupplyDenom: "apmt",
	}
	d := buildBondedReport(chain)
	if d.BondedPct != 25 {
		t.Fatalf("BondedPct = %v, want 25 from total supply", d.BondedPct)
	}
	if d.MintBondedPct != 50 {
		t.Fatalf("MintBondedPct = %v, want 50 from staking pool", d.MintBondedPct)
	}
}
