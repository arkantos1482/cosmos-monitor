package fetchall

import (
	"sync"

	"github.com/arkantos1482/cosmos-monitor/internal/fetch"
)

// Snapshots is the result of a parallel fetch from all data sources.
type Snapshots struct {
	Chain  fetch.ChainSnapshot
	EVM    fetch.EVMSnapshot
	System fetch.SystemSnapshot
	Docker fetch.DockerSnapshot
}

// Load fetches chain, EVM, system, and Docker metrics concurrently.
func Load(rpc, rest, evm, container string) Snapshots {
	var (
		chain  fetch.ChainSnapshot
		evSnap fetch.EVMSnapshot
		sys    fetch.SystemSnapshot
		docker fetch.DockerSnapshot
		params fetch.ChainParams
		wg     sync.WaitGroup
	)
	wg.Add(5)
	go func() { defer wg.Done(); chain = fetch.FetchChain(rpc, rest) }()
	go func() { defer wg.Done(); evSnap = fetch.FetchEVM(evm) }()
	go func() { defer wg.Done(); sys = fetch.FetchSystem() }()
	go func() { defer wg.Done(); docker = fetch.FetchDocker(container) }()
	go func() { defer wg.Done(); params = fetch.FetchParams(rest) }()
	wg.Wait()
	chain.Params = params
	return Snapshots{Chain: chain, EVM: evSnap, System: sys, Docker: docker}
}
