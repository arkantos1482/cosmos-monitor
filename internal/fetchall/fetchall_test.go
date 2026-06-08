package fetchall

import (
	"testing"

	"github.com/arkantos1482/cosmos-monitor/internal/panel"
)

func TestLoadForReturnsViewScopedSnapshots(t *testing.T) {
	infra := LoadFor(panel.ViewInfra, "http://127.0.0.1:1", "http://127.0.0.1:1", "http://127.0.0.1:1", "none")
	if infra.Chain.BlockHeight != 0 {
		t.Fatal("infra view should not fetch chain block height")
	}
}

func TestChainOptsForGovernanceSkipsEconomics(t *testing.T) {
	o := chainOptsFor(panel.ViewGovernance)
	if !o.SkipValidatorRewards || !o.SkipEconomics || o.SkipGovernance {
		t.Fatalf("unexpected opts: %+v", o)
	}
}
