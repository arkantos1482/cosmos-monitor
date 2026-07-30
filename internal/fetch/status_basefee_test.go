package fetch

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFetchChainStatusBaseFeeRESTFallbackToBlockResults(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/status", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"result": map[string]any{
				"node_info": map[string]any{"moniker": "n1"},
				"sync_info": map[string]any{
					"latest_block_height": "42",
					"latest_block_time":   "2026-07-29T00:00:00Z",
					"catching_up":         false,
				},
			},
		})
	})
	mux.HandleFunc("/net_info", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"result": map[string]any{"n_peers": "3"},
		})
	})
	mux.HandleFunc("/block_results", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("height") != "42" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"result": map[string]any{
				"height": "42",
				"finalize_block_events": []map[string]any{
					{
						"type": "fee_market",
						"attributes": []map[string]string{
							{"key": "base_fee", "value": "0.000000000000000007"},
							{"key": "mode", "value": "BeginBlock"},
						},
					},
				},
			},
		})
	})
	rpc := httptest.NewServer(mux)
	defer rpc.Close()

	rest := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "api disabled", http.StatusBadGateway)
	}))
	defer rest.Close()

	snap := FetchChainStatus(rpc.URL, rest.URL)
	if snap.Err != nil {
		t.Fatalf("unexpected err: %v", snap.Err)
	}
	if snap.BaseFee != "0.000000000000000007" {
		t.Fatalf("expected block_results base fee fallback, got %q", snap.BaseFee)
	}
	if snap.PeerCount != 3 {
		t.Fatalf("peers=%d", snap.PeerCount)
	}
}

func TestFetchChainStatusPrefersRESTBaseFee(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/status", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"result": map[string]any{
				"node_info": map[string]any{"moniker": "n1"},
				"sync_info": map[string]any{
					"latest_block_height": "10",
					"latest_block_time":   "2026-07-29T00:00:00Z",
					"catching_up":         false,
				},
			},
		})
	})
	mux.HandleFunc("/net_info", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"n_peers": "1"}})
	})
	mux.HandleFunc("/block_results", func(w http.ResponseWriter, _ *http.Request) {
		t.Fatal("block_results should not be required when REST base_fee works")
	})
	rpc := httptest.NewServer(mux)
	defer rpc.Close()

	rest := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/cosmos/evm/feemarket/v1/base_fee") {
			fmt.Fprint(w, `{"base_fee":"7"}`)
			return
		}
		http.NotFound(w, r)
	}))
	defer rest.Close()

	snap := FetchChainStatus(rpc.URL, rest.URL)
	if snap.BaseFee != "7" {
		t.Fatalf("got %q", snap.BaseFee)
	}
}
