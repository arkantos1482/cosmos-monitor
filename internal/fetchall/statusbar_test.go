package fetchall

import (
	"testing"
	"time"

	"github.com/arkantos1482/cosmos-monitor/internal/fetch"
	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func TestMergeStatusOverlayFillsMissingEVMAndDocker(t *testing.T) {
	view := Snapshots{
		Chain: fetch.ChainSnapshot{
			BlockHeight: 100,
			PeerCount:   3,
			Moniker:     "node1",
		},
	}
	bar := Snapshots{
		Chain:  fetch.ChainSnapshot{BlockHeight: 101, PeerCount: 4},
		EVM:    fetch.EVMSnapshot{PeerCount: 2},
		Docker: fetch.DockerSnapshot{Running: true},
	}
	got := mergeStatusOverlay(view, bar, model.StatusAvailability{ChainOK: true, EVMOK: true, DockerOK: true})

	if got.Chain.BlockHeight != 101 {
		t.Fatalf("height = %d, want 101", got.Chain.BlockHeight)
	}
	if got.Chain.PeerCount != 4 {
		t.Fatalf("cosmos peers = %d, want 4", got.Chain.PeerCount)
	}
	if got.EVM.PeerCount != 2 {
		t.Fatalf("evm peers = %d, want 2", got.EVM.PeerCount)
	}
	if !got.Docker.Running {
		t.Fatal("expected docker running")
	}
	if got.Chain.Moniker != "node1" {
		t.Fatalf("moniker = %q, want node1", got.Chain.Moniker)
	}
}

func TestMergeStatusOverlaySkipsFailedBarSources(t *testing.T) {
	view := Snapshots{
		Chain: fetch.ChainSnapshot{BlockHeight: 50, PeerCount: 1},
	}
	bar := Snapshots{
		EVM:    fetch.EVMSnapshot{PeerCount: 9},
		Docker: fetch.DockerSnapshot{Running: true},
	}
	got := mergeStatusOverlay(view, bar, model.StatusAvailability{ChainOK: false, EVMOK: false, DockerOK: false})

	if got.Chain.BlockHeight != 50 || got.Chain.PeerCount != 1 {
		t.Fatalf("view chain changed: %+v", got.Chain)
	}
	if got.EVM.PeerCount != 0 || got.Docker.Running {
		t.Fatalf("unexpected overlay from failed bar: evm=%d docker=%v", got.EVM.PeerCount, got.Docker.Running)
	}
}

func TestStatusBarCacheHit(t *testing.T) {
	statusBarCache.mu.Lock()
	statusBarCache.key = statusBarCacheKey{rpc: "r", rest: "s", evm: "e", container: "c"}
	statusBarCache.snap = Snapshots{Chain: fetch.ChainSnapshot{BlockHeight: 42}}
	statusBarCache.bar = model.StatusAvailability{ChainOK: true}
	statusBarCache.at = time.Now()
	statusBarCache.mu.Unlock()

	snap, bar := fetchStatusBar("r", "s", "e", "c")
	if snap.Chain.BlockHeight != 42 || !bar.ChainOK {
		t.Fatalf("cache miss: %+v %+v", snap, bar)
	}
}
