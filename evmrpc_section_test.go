package main

import (
	"strings"
	"testing"
)

func TestBuildMarkdownEVMRPCWebStrip(t *testing.T) {
	d := WebData{
		EVMRPCOk: true, EVMSynced: true, EVMListening: true,
		EVMBlockAge: "4.2s", EVMBlock: "100", EVMChainID: 290290,
		Network: "pmt", EVMHTTPEndpoint: "http://localhost:8545",
		RPCProbeOK: 9, RPCProbeTotal: 9,
		RPCProbes: []WebRPCProbe{
			{Method: "eth_blockNumber", OK: true, Latency: "12ms", Request: `{"jsonrpc":"2.0"}`, Response: `{}`},
		},
		GasPrice: "1 apmt", PendingTx: 2, QueuedTx: 1,
	}
	md := buildMarkdown(d, true)
	if !strings.Contains(md, `class="evm-rpc-strip"`) {
		t.Fatal("web markdown should include EVM status strip")
	}
	if !strings.Contains(md, "## For operators") {
		t.Fatal("expected For operators subsection")
	}
	if !strings.Contains(md, "## Live (JSON-RPC)") {
		t.Fatal("expected Live subsection")
	}
	if !strings.Contains(md, "## Probe health") {
		t.Fatal("expected Probe health subsection")
	}
	if strings.Contains(md, "## EVM Config") {
		t.Fatal("EVM Config should be removed from section 7")
	}
	if strings.Contains(md, "Raw JSON-RPC Samples") {
		t.Fatal("raw samples section should be removed")
	}
}

func TestRenderFragmentEVMProbeTable(t *testing.T) {
	d := WebData{
		EVMRPCOk: false, EVMSynced: true,
		EVMHTTPEndpoint: "http://127.0.0.1:8545",
		RPCProbeOK: 8, RPCProbeTotal: 9,
		RPCProbes: []WebRPCProbe{
			{Method: "eth_chainId", OK: true, Latency: "5ms", Request: `{}`, Response: `{}`},
			{Method: "txpool_status", OK: false, Latency: "1ms", Error: "connection refused",
				Request: `{"jsonrpc":"2.0","method":"txpool_status","params":[],"id":1}`,
				Response: `{}`},
		},
	}
	out := renderFragment(d)
	if !strings.Contains(out, `class="evm-probe-table"`) {
		t.Fatal("fragment should include probe table")
	}
	if !strings.Contains(out, `<details class="evm-probe-detail"`) {
		t.Fatal("failed probe should expand in details")
	}
	if !strings.Contains(out, "curl -sS") {
		t.Fatal("failed probe should include curl command")
	}
}

func TestSortedRPCProbesGroupsAndFailuresFirst(t *testing.T) {
	probes := []WebRPCProbe{
		{Method: "net_listening", OK: true},
		{Method: "eth_syncing", OK: false},
		{Method: "eth_blockNumber", OK: true},
	}
	got := sortedRPCProbes(probes)
	if got[0].Method != "eth_syncing" {
		t.Fatalf("expected eth failure first, got %s", got[0].Method)
	}
	if probeNamespace(got[len(got)-1].Method) != "net" {
		t.Fatalf("expected net last, got %s", got[len(got)-1].Method)
	}
}

func TestEvmWSEndpoint(t *testing.T) {
	if ws := evmWSEndpoint("http://localhost:8545"); ws != "ws://localhost:8546" {
		t.Fatalf("unexpected ws endpoint: %s", ws)
	}
}
