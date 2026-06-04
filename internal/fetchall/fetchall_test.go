package fetchall

import (
	"testing"
	"time"

	"github.com/arkantos1482/cosmos-monitor/internal/panel"
)

func TestLoadForUsesSnapshotCache(t *testing.T) {
	cache.mu.Lock()
	cache.snapAt = time.Now()
	cache.snap = Snapshots{}
	cache.mu.Unlock()

	start := time.Now()
	_ = LoadFor(panel.ViewInfra, "http://127.0.0.1:1", "http://127.0.0.1:1", "http://127.0.0.1:1", "none")
	if time.Since(start) > 200*time.Millisecond {
		t.Fatalf("expected cache hit to return quickly, took %v", time.Since(start))
	}
}

func TestChainOptsForGovernanceSkipsEconomics(t *testing.T) {
	o := chainOptsFor(panel.ViewGovernance)
	if !o.SkipValidatorRewards || !o.SkipEconomics || o.SkipGovernance {
		t.Fatalf("unexpected opts: %+v", o)
	}
}
