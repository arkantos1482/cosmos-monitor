package fetchall

import (
	"testing"
	"time"

	"github.com/arkantos1482/cosmos-monitor/internal/fetch"
	"github.com/arkantos1482/cosmos-monitor/internal/panel"
)

func TestLoadForReturnsViewScopedSnapshots(t *testing.T) {
	infra := LoadFor(panel.ViewInfra, "http://127.0.0.1:1", "http://127.0.0.1:1", "http://127.0.0.1:1", "none")
	if infra.Chain.BlockHeight != 0 {
		t.Fatal("infra view should not fetch chain block height")
	}
}

func TestLoadForCachesPerView(t *testing.T) {
	const dead = "http://127.0.0.1:1"
	key := viewCacheKey{view: panel.ViewInfra, rpc: dead, rest: dead, evm: dead, container: "none"}

	LoadFor(panel.ViewInfra, dead, dead, dead, "none")
	cache.mu.Lock()
	at1 := cache.byView[key].at
	cache.mu.Unlock()

	time.Sleep(10 * time.Millisecond)
	LoadFor(panel.ViewInfra, dead, dead, dead, "none")
	cache.mu.Lock()
	at2 := cache.byView[key].at
	cache.mu.Unlock()
	if !at1.Equal(at2) {
		t.Fatal("second LoadFor within TTL should reuse cached snapshot")
	}

	LoadFor(panel.ViewEVM, dead, dead, dead, "none")
	cache.mu.Lock()
	n := len(cache.byView)
	cache.mu.Unlock()
	if n < 2 {
		t.Fatal("different views should have separate cache entries")
	}
}

func TestLoadForCacheExpires(t *testing.T) {
	const dead = "http://127.0.0.1:1"
	key := viewCacheKey{view: panel.ViewInfra, rpc: dead, rest: dead, evm: dead, container: "none"}
	cache.mu.Lock()
	cache.byView = map[viewCacheKey]cachedSnapshot{
		key: {snap: Snapshots{System: fetch.SystemSnapshot{LoadAvg1: 999.99}}, at: time.Now().Add(-snapshotTTL)},
	}
	cache.mu.Unlock()

	snap := LoadFor(panel.ViewInfra, dead, dead, dead, "none")
	if snap.System.LoadAvg1 == 999.99 {
		t.Fatal("expired cache entry should be refreshed")
	}
}

func TestChainRecipeForGovernanceSkipsEconomics(t *testing.T) {
	r := chainRecipeFor(panel.ViewGovernance)
	if r.ValidatorRewards || r.MintData || r.FeemarketLive || r.Governance == false || !r.ModuleBalances {
		t.Fatalf("unexpected recipe: %+v", r)
	}
}

func TestChainRecipeForNodeSkipsUnusedFetches(t *testing.T) {
	r := chainRecipeFor(panel.ViewNode)
	if r.ValidatorRewards || r.MintData || r.Governance || r.ConsensusParams ||
		r.LocalStaking || r.ValidatorScope != fetch.ValidatorsBonded {
		t.Fatalf("unexpected recipe: %+v", r)
	}
	if !r.CometExtended || !r.StakingPool || !r.SigningInfos {
		t.Fatalf("node recipe missing required fetches: %+v", r)
	}
}

func TestChainRecipeForDistributionFetchesValidatorRewards(t *testing.T) {
	r := chainRecipeFor(panel.ViewDistribution)
	if !r.ValidatorRewards || r.Governance || !r.ModuleBalances || !r.CommunityPool {
		t.Fatalf("unexpected recipe: %+v", r)
	}
}
