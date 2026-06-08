package fetchall

import (
	"sync"
	"time"

	"github.com/arkantos1482/cosmos-monitor/internal/fetch"
	"github.com/arkantos1482/cosmos-monitor/internal/panel"
)

// Snapshots is the result of a parallel fetch from all data sources.
type Snapshots struct {
	Chain  fetch.ChainSnapshot
	EVM    fetch.EVMSnapshot
	System fetch.SystemSnapshot
	Docker fetch.DockerSnapshot
}

const paramsTTL = 5 * time.Minute

var cache struct {
	mu          sync.Mutex
	params      fetch.ChainParams
	paramsAt    time.Time
	lastMoniker string
}

// Load fetches all sources (full dashboard refresh).
func Load(rpc, rest, evm, container string) Snapshots {
	return LoadFor(panel.ViewHome, rpc, rest, evm, container)
}

// LoadFor fetches what the active section needs (view-scoped, no snapshot cache).
func LoadFor(view panel.View, rpc, rest, evm, container string) Snapshots {
	snap := fetchForView(view, rpc, rest, evm, container)
	if snap.Chain.Moniker != "" {
		cache.mu.Lock()
		cache.lastMoniker = snap.Chain.Moniker
		cache.mu.Unlock()
	}
	return snap
}

// Moniker returns the latest known node moniker (for page headers on lightweight fetches).
func Moniker() string {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	return cache.lastMoniker
}

func cachedParams(rest string) fetch.ChainParams {
	cache.mu.Lock()
	if !cache.paramsAt.IsZero() && time.Since(cache.paramsAt) < paramsTTL {
		p := cache.params
		cache.mu.Unlock()
		return p
	}
	cache.mu.Unlock()

	p := fetch.FetchParams(rest)

	cache.mu.Lock()
	cache.params = p
	cache.paramsAt = time.Now()
	cache.mu.Unlock()
	return p
}

func fetchForView(view panel.View, rpc, rest, evm, container string) Snapshots {
	switch view {
	case panel.ViewInfra:
		var sys fetch.SystemSnapshot
		var docker fetch.DockerSnapshot
		var wg sync.WaitGroup
		wg.Add(2)
		go func() { defer wg.Done(); sys = fetch.FetchSystem() }()
		go func() { defer wg.Done(); docker = fetch.FetchDocker(container) }()
		wg.Wait()
		return Snapshots{System: sys, Docker: docker}

	case panel.ViewEVM:
		var evSnap fetch.EVMSnapshot
		var p fetch.ChainParams
		var wg sync.WaitGroup
		wg.Add(2)
		go func() { defer wg.Done(); evSnap = fetch.FetchEVM(evm) }()
		go func() { defer wg.Done(); p = cachedParams(rest) }()
		wg.Wait()
		return Snapshots{EVM: evSnap, Chain: fetch.ChainSnapshot{Params: p}}
	}

	chainOpts := chainOptsFor(view)
	needEVM := view == panel.ViewHome || view == panel.ViewNode
	needSys := view == panel.ViewHome
	needDocker := view == panel.ViewHome

	var (
		chain  fetch.ChainSnapshot
		evSnap fetch.EVMSnapshot
		sys    fetch.SystemSnapshot
		docker fetch.DockerSnapshot
		p      fetch.ChainParams
		wg     sync.WaitGroup
	)
	wg.Add(2)
	go func() { defer wg.Done(); chain = fetch.FetchChain(rpc, rest, chainOpts) }()
	go func() { defer wg.Done(); p = cachedParams(rest) }()
	if needEVM {
		wg.Add(1)
		go func() { defer wg.Done(); evSnap = fetch.FetchEVM(evm) }()
	}
	if needSys {
		wg.Add(1)
		go func() { defer wg.Done(); sys = fetch.FetchSystem() }()
	}
	if needDocker {
		wg.Add(1)
		go func() { defer wg.Done(); docker = fetch.FetchDocker(container) }()
	}
	wg.Wait()
	chain.Params = p
	return Snapshots{Chain: chain, EVM: evSnap, System: sys, Docker: docker}
}

func chainOptsFor(view panel.View) fetch.ChainOpts {
	switch view {
	case panel.ViewNode, panel.ViewLocalValidator:
		return fetch.ChainOpts{
			SkipValidatorRewards: true,
			SkipGovernance:       true,
			SkipEconomics:        true,
		}
	case panel.ViewGovernance:
		return fetch.ChainOpts{
			SkipValidatorRewards: true,
			SkipEconomics:        true,
		}
	default:
		return fetch.ChainOpts{}
	}
}
