package fetch

import (
	"encoding/json"
	"testing"
)

func TestParseBlockResultsGasUsedAndEvent(t *testing.T) {
	raw := blockResultsResp{}
	err := json.Unmarshal([]byte(`{
		"result": {
			"height": "100",
			"txs_results": [
				{"code": 0, "gas_wanted": "25000", "gas_used": "21000", "log": ""},
				{"code": 0, "gas_wanted": "8000", "gas_used": "5000", "log": ""}
			],
			"end_block_events": [
				{
					"type": "block_gas",
					"attributes": [
						{"key": "height", "value": "100"},
						{"key": "amount", "value": "26000"}
					]
				}
			],
			"begin_block_events": [
				{
					"type": "fee_market",
					"attributes": [{"key": "base_fee", "value": "1000000000000"}]
				}
			]
		}
	}`), &raw)
	if err != nil {
		t.Fatal(err)
	}
	s := ParseBlockResults(raw)
	if s.GasUsedSum != 26000 {
		t.Fatalf("gas used sum: got %d want 26000", s.GasUsedSum)
	}
	if s.TxGasWantedSum != 33000 {
		t.Fatalf("tx gas wanted sum: got %d want 33000", s.TxGasWantedSum)
	}
	if s.BlockGasWanted != 26000 {
		t.Fatalf("block gas wanted: got %d want 26000", s.BlockGasWanted)
	}
	if s.BaseFeeEvent != "1000000000000" {
		t.Fatalf("base fee event: %q", s.BaseFeeEvent)
	}
}

func TestParseBlockResultsFinalizeBlockEvents(t *testing.T) {
	raw := blockResultsResp{}
	err := json.Unmarshal([]byte(`{
		"result": {
			"height": "649160",
			"txs_results": null,
			"finalize_block_events": [
				{
					"type": "fee_market",
					"attributes": [
						{"key": "base_fee", "value": "7"},
						{"key": "mode", "value": "BeginBlock"}
					]
				},
				{
					"type": "block_gas",
					"attributes": [
						{"key": "height", "value": "649160"},
						{"key": "amount", "value": "0"},
						{"key": "mode", "value": "EndBlock"}
					]
				}
			]
		}
	}`), &raw)
	if err != nil {
		t.Fatal(err)
	}
	s := ParseBlockResults(raw)
	if s.BlockGasWanted != 0 {
		t.Fatalf("block gas wanted: got %d want 0", s.BlockGasWanted)
	}
	if s.BaseFeeEvent != "7" {
		t.Fatalf("base fee event: %q", s.BaseFeeEvent)
	}
}

func TestShouldIgnoreGasUsedStopsSum(t *testing.T) {
	raw := blockResultsResp{}
	_ = json.Unmarshal([]byte(`{
		"result": {
			"txs_results": [
				{"code": 0, "gas_used": "1000", "log": ""},
				{"code": 11, "gas_used": "999", "log": "no block gas left to run tx: out of gas"},
				{"code": 0, "gas_used": "500", "log": ""}
			]
		}
	}`), &raw)
	s := ParseBlockResults(raw)
	if s.GasUsedSum != 1000 {
		t.Fatalf("expected 1000 before break, got %d", s.GasUsedSum)
	}
}
