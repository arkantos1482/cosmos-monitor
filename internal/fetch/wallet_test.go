package fetch

import (
	"encoding/json"
	"testing"
)

func TestEVMParamsUnmarshalHistoryServeWindowString(t *testing.T) {
	const raw = `{"params":{"evm_denom":"apmt","history_serve_window":"8192"}}`
	var ep evmParamsResp
	if err := json.Unmarshal([]byte(raw), &ep); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if ep.Params.EvmDenom != "apmt" {
		t.Fatalf("evm_denom = %q, want apmt", ep.Params.EvmDenom)
	}
	if ep.Params.HistoryServeWindow != "8192" {
		t.Fatalf("history_serve_window = %q, want 8192", ep.Params.HistoryServeWindow)
	}
}

func TestBankDenomMetadataUnmarshal(t *testing.T) {
	const raw = `{"metadata":{"name":"PMT","symbol":"PMT","display":"pmt","base":"apmt","denom_units":[{"denom":"apmt","exponent":0},{"denom":"pmt","exponent":18}]}}`
	var meta bankDenomMetadataResp
	if err := json.Unmarshal([]byte(raw), &meta); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got := evmDisplayDecimals(meta.Metadata.Display, meta.Metadata.DenomUnits); got != 18 {
		t.Fatalf("decimals = %d, want 18", got)
	}
}
