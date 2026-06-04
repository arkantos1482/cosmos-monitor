package main

import (
	"strings"
	"testing"
)

func TestBuildMarkdownEVMRPCStrip(t *testing.T) {
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
	md := buildMarkdown(d)
	if !strings.Contains(md, "**RPC: OK**") {
		t.Fatal("markdown should include RPC status line")
	}
	if strings.Contains(md, `class="evm-rpc-strip"`) {
		t.Fatal("markdown should not use HTML pill strip")
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
}

func TestRenderFragmentEVMProbeLog(t *testing.T) {
	d := WebData{
		EVMRPCOk: false, EVMSynced: true,
		EVMHTTPEndpoint: "http://127.0.0.1:8545",
		RPCProbeOK: 8, RPCProbeTotal: 9,
		RPCProbes: []WebRPCProbe{
			{Method: "eth_chainId", OK: true, Latency: "5ms",
				Request: `{"jsonrpc":"2.0","method":"eth_chainId","params":[],"id":1}`,
				Response: `{"jsonrpc":"2.0","id":1,"result":"0x46f92"}`},
			{Method: "txpool_status", OK: false, Latency: "1ms", Error: "connection refused",
				Request: `{"jsonrpc":"2.0","method":"txpool_status","params":[],"id":1}`,
				Response: `{"jsonrpc":"2.0","id":1,"error":{"message":"connection refused"}}`},
		},
	}
	out := renderFragment(d)
	if !strings.Contains(out, "[ETH]") || !strings.Contains(out, "eth_chainId") {
		t.Fatal("probe log should list eth namespace and methods")
	}
	if !strings.Contains(out, "res »") || !strings.Contains(out, "0x46f92") {
		t.Fatal("each probe should show JSON response body")
	}
	if !strings.Contains(out, "curl -sS") {
		t.Fatal("failed probe should include curl in bash fence")
	}
}

func TestFormatProbeExchange(t *testing.T) {
	body := formatProbeExchange(WebRPCProbe{
		Method: "eth_blockNumber", OK: true, Latency: "3ms",
		Request:  `{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}`,
		Response: `{"jsonrpc":"2.0","id":1,"result":"0x10"}`,
	})
	if !strings.Contains(body, "req »") || !strings.Contains(body, "res »") {
		t.Fatalf("expected req/res markers: %q", body)
	}
	if !strings.Contains(body, `"result": "0x10"`) {
		t.Fatalf("expected indented result: %q", body)
	}
}

func TestRenderProbeLogFormat(t *testing.T) {
	log := renderProbeLog([]WebRPCProbe{
		{Method: "eth_blockNumber", OK: true, Latency: "3ms"},
		{Method: "net_listening", OK: true, Latency: "1ms"},
	})
	if !strings.Contains(log, "[ETH]") || !strings.Contains(log, "[NET]") {
		t.Fatalf("expected namespace headers: %q", log)
	}
	if !strings.Contains(log, "·  eth_blockNumber") {
		t.Fatalf("expected ok marker line: %q", log)
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
