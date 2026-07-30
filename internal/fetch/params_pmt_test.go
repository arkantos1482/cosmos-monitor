package fetch

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchParamsPMTRewardsOK(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/cosmos/evm/pmtrewards/v1/params", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"params": map[string]any{
				"enabled": true,
				"reward_per_block": map[string]string{
					"denom":  "apmt",
					"amount": "100000000000000000",
				},
				"pool_address": "cosmos1pool",
			},
		})
	})
	mux.HandleFunc("/cosmos/bank/v1beta1/balances/cosmos1pool", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"balances": []map[string]string{},
		})
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	p := FetchParams(srv.URL)
	if !p.PMTRewardsParamsOK {
		t.Fatal("expected PMTRewardsParamsOK")
	}
	if !p.PMTRewardsEnabled {
		t.Fatal("expected enabled")
	}
	if !p.PMTRewardsPoolBalanceOK {
		t.Fatal("expected pool balance fetch OK")
	}
	if p.PMTRewardsPoolBalanceAmt != "" {
		t.Fatalf("empty pool should leave amount empty, got %q", p.PMTRewardsPoolBalanceAmt)
	}
}

func TestFetchParamsPMTRewardsMissing(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"code":12,"message":"Not Implemented"}`))
	}))
	defer srv.Close()

	p := FetchParams(srv.URL)
	if p.PMTRewardsParamsOK {
		t.Fatal("params must not be OK when REST fails/unimplemented")
	}
	if p.PMTRewardsEnabled {
		t.Fatal("enabled must stay false when params not fetched")
	}
}
