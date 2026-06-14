package fetchall

import (
	"sync"
	"time"

	"github.com/arkantos1482/cosmos-monitor/internal/fetch"
	"github.com/arkantos1482/cosmos-monitor/internal/model"
	"github.com/arkantos1482/cosmos-monitor/internal/panel"
)

type statusBarCacheKey struct {
	view                        panel.View
	rpc, rest, evm, container string
}

var statusBarCache struct {
	mu   sync.Mutex
	key  statusBarCacheKey
	snap Snapshots
	bar  model.StatusAvailability
	at   time.Time
}

func fetchStatusBar(view panel.View, rpc, rest, evm, container string) (Snapshots, model.StatusAvailability) {
	key := statusBarCacheKey{view: view, rpc: rpc, rest: rest, evm: evm, container: container}
	skipEVMPeer := view == panel.ViewHome || view == panel.ViewEVM

	statusBarCache.mu.Lock()
	if !statusBarCache.at.IsZero() &&
		statusBarCache.key == key &&
		time.Since(statusBarCache.at) < snapshotTTL {
		snap := statusBarCache.snap
		bar := statusBarCache.bar
		statusBarCache.mu.Unlock()
		return snap, bar
	}
	statusBarCache.mu.Unlock()

	var (
		chain  fetch.ChainSnapshot
		evSnap fetch.EVMSnapshot
		docker fetch.DockerSnapshot
		p      fetch.ChainParams
		wg     sync.WaitGroup
	)
	wg.Add(2)
	go func() { defer wg.Done(); chain = fetch.FetchChainStatus(rpc, rest) }()
	go func() { defer wg.Done(); docker = fetch.FetchDockerRunning(container) }()
	if view != panel.ViewEVM {
		wg.Add(1)
		go func() { defer wg.Done(); p = cachedParams(rest) }()
	}
	if !skipEVMPeer {
		wg.Add(1)
		go func() { defer wg.Done(); evSnap = fetch.FetchEVMPeerCount(evm) }()
	}
	wg.Wait()
	chain.Params = p

	snap := Snapshots{Chain: chain, EVM: evSnap, Docker: docker}
	bar := model.StatusAvailability{
		ChainOK:  chain.Err == nil,
		EVMOK:    !skipEVMPeer && evSnap.Err == nil,
		DockerOK: docker.Err == nil,
	}

	statusBarCache.mu.Lock()
	statusBarCache.key = key
	statusBarCache.snap = snap
	statusBarCache.bar = bar
	statusBarCache.at = time.Now()
	statusBarCache.mu.Unlock()

	return snap, bar
}

func mergeStatusOverlay(view, bar Snapshots, barOK model.StatusAvailability) Snapshots {
	if barOK.ChainOK {
		view.Chain.BlockHeight = bar.Chain.BlockHeight
		view.Chain.CatchingUp = bar.Chain.CatchingUp
		view.Chain.LatestBlockTime = bar.Chain.LatestBlockTime
		view.Chain.PeerCount = bar.Chain.PeerCount
		if bar.Chain.BaseFee != "" {
			view.Chain.BaseFee = bar.Chain.BaseFee
		}
		if view.Chain.Moniker == "" {
			view.Chain.Moniker = bar.Chain.Moniker
		}
	}
	if barOK.EVMOK {
		view.EVM.PeerCount = bar.EVM.PeerCount
	}
	if barOK.DockerOK {
		view.Docker.Running = bar.Docker.Running
	}
	return view
}
