package fetch

import (
	"testing"
)

func TestFetchChainStatusEmptyRPC(t *testing.T) {
	snap := FetchChainStatus("http://127.0.0.1:1", "http://127.0.0.1:1")
	if snap.Err == nil {
		t.Fatal("expected error for unreachable RPC")
	}
}

func TestFetchEVMPeerCountEmptyRPC(t *testing.T) {
	snap := FetchEVMPeerCount("http://127.0.0.1:1")
	if snap.Err == nil {
		t.Fatal("expected error for unreachable EVM RPC")
	}
}

func TestFetchDockerRunningMissingContainer(t *testing.T) {
	snap := FetchDockerRunning("nonexistent-pmtop-test-container")
	if snap.Err == nil {
		t.Fatal("expected error for missing container")
	}
}
