package fetch

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFetchEVMBlockNumberFailKeepsListening(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Method string `json:"method"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		w.Header().Set("Content-Type", "application/json")
		switch req.Method {
		case "eth_blockNumber":
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"error":{"code":-32000,"message":"blockNumber down"}}`))
		case "net_listening":
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":true}`))
		default:
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":null}`))
		}
	}))
	defer srv.Close()

	snap := FetchEVM(srv.URL)
	if snap.Err == nil {
		t.Fatal("expected eth_blockNumber error")
	}
	if !strings.Contains(snap.Err.Error(), "eth_blockNumber") {
		t.Fatalf("unexpected err: %v", snap.Err)
	}
	if !snap.HasNetListening {
		t.Fatal("net_listening must still be recorded when blockNumber fails")
	}
	if !snap.NetListening {
		t.Fatal("expected NetListening true from successful net_listening probe")
	}
}

func TestFetchEVMListeningUnknownWhenProbeFails(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Method string `json:"method"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		w.Header().Set("Content-Type", "application/json")
		switch req.Method {
		case "eth_blockNumber":
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"0x10"}`))
		case "net_listening":
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"error":{"code":-32601,"message":"method not found"}}`))
		default:
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":false}`))
		}
	}))
	defer srv.Close()

	snap := FetchEVM(srv.URL)
	if snap.Err != nil {
		t.Fatalf("unexpected err: %v", snap.Err)
	}
	if snap.HasNetListening {
		t.Fatal("HasNetListening must be false when probe fails")
	}
	if snap.NetListening {
		t.Fatal("NetListening must stay false when unknown")
	}
}
