package report

import (
	"testing"

	"github.com/arkantos1482/cosmos-monitor/internal/fetch"
	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func TestBuildPMTParamsMissingNotDisabled(t *testing.T) {
	d := Build(fetch.ChainSnapshot{}, fetch.EVMSnapshot{}, fetch.SystemSnapshot{}, fetch.DockerSnapshot{},
		"", model.StatusAvailability{ChainOK: true}, fetch.AppTomlGasConfig{}, nil)
	if d.HasPMTParams {
		t.Fatal("expected HasPMTParams false")
	}
	if d.PMTEnabled {
		t.Fatal("must not mark enabled when params missing")
	}
	if d.PMTPoolEmpty {
		t.Fatal("must not mark pool empty when balance was not fetched")
	}
}

func TestBuildPMTEnabledPoolEmpty(t *testing.T) {
	d := Build(fetch.ChainSnapshot{
		Params: fetch.ChainParams{
			PMTRewardsParamsOK:       true,
			PMTRewardsEnabled:        true,
			PMTRewardsPoolBalanceOK:  true,
			PMTRewardsPoolAddress:    "cosmos1pool",
			RewardPerBlockAmount:     "100000000000000000",
			RewardPerBlockDenom:      "apmt",
			PMTRewardsPoolBalanceAmt: "0",
			BondDenom:                "apmt",
		},
	}, fetch.EVMSnapshot{}, fetch.SystemSnapshot{}, fetch.DockerSnapshot{},
		"", model.StatusAvailability{ChainOK: true}, fetch.AppTomlGasConfig{}, nil)
	if !d.HasPMTParams || !d.PMTEnabled || !d.PMTPoolEmpty {
		t.Fatalf("got Has=%v Enabled=%v Empty=%v", d.HasPMTParams, d.PMTEnabled, d.PMTPoolEmpty)
	}
}
