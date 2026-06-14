package report

import (
	"testing"

	"github.com/arkantos1482/cosmos-monitor/internal/fetch"
	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func buildBondedReport(chain fetch.ChainSnapshot) model.Report {
	return Build(chain, fetch.EVMSnapshot{}, fetch.SystemSnapshot{}, fetch.DockerSnapshot{}, "", model.StatusAvailability{}, fetch.AppTomlGasConfig{}, nil)
}

func TestBondedPctFromStakingPool(t *testing.T) {
	chain := fetch.ChainSnapshot{
		Params:          fetch.ChainParams{BondDenom: "apmt"},
		BondedTokens:    "4000000000000000000000000",
		NotBondedTokens: "8000000000000000000000000",
	}
	d := buildBondedReport(chain)

	const want = 33.333333333333336
	if d.BondedPct < want-0.01 || d.BondedPct > want+0.01 {
		t.Fatalf("BondedPct = %v, want ~33.33", d.BondedPct)
	}
}

func TestBondedPctFromStakingPoolIgnoresBankSupply(t *testing.T) {
	chain := fetch.ChainSnapshot{
		Params:           fetch.ChainParams{BondDenom: "apmt"},
		BondedTokens:     "1000000000000000000000000",
		NotBondedTokens:  "1000000000000000000000000",
		TotalSupply:      "4000000000000000000000000",
		TotalSupplyDenom: "apmt",
	}
	d := buildBondedReport(chain)
	if d.BondedPct != 50 {
		t.Fatalf("BondedPct = %v, want 50 from staking pool (not bank supply)", d.BondedPct)
	}
}
