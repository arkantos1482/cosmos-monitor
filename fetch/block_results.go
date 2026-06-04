package fetch

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
)

type blockResultsResp struct {
	Result struct {
		Height           string `json:"height"`
		TxsResults       []txResult `json:"txs_results"`
		EndBlockEvents   []abciEvent `json:"end_block_events"`
		BeginBlockEvents []abciEvent `json:"begin_block_events"`
	} `json:"result"`
}

type txResult struct {
	Code    int    `json:"code"`
	GasUsed string `json:"gas_used"`
	Log     string `json:"log"`
}

type abciEvent struct {
	Type       string         `json:"type"`
	Attributes []abciAttribute `json:"attributes"`
}

type abciAttribute struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// BlockResultsSummary holds gas and feemarket data parsed from CometBFT block_results.
type BlockResultsSummary struct {
	GasUsedSum      uint64
	BlockGasWanted  uint64 // from block_gas end_block event
	BaseFeeEvent    string // from fee_market begin_block event (raw)
	OK              bool
}

func attrValue(attrs []abciAttribute, key string) string {
	for _, a := range attrs {
		k := decodeABCIString(a.Key)
		if k == key {
			return decodeABCIString(a.Value)
		}
	}
	return ""
}

func decodeABCIString(s string) string {
	if s == "" {
		return ""
	}
	if strings.ContainsAny(s, " \n\t") {
		return s
	}
	// CometBFT JSON may base64-encode attribute keys/values.
	if b, err := base64.StdEncoding.DecodeString(s); err == nil && len(b) > 0 && isPrintableASCII(b) {
		return string(b)
	}
	return s
}

func isPrintableASCII(b []byte) bool {
	for _, c := range b {
		if c < 0x20 || c > 0x7e {
			return false
		}
	}
	return true
}

func shouldIgnoreGasUsed(res txResult) bool {
	return res.Code == 11 && strings.Contains(res.Log, "no block gas left to run tx: out of gas")
}

func sumBlockGasUsed(txs []txResult) uint64 {
	var sum uint64
	for _, tx := range txs {
		if shouldIgnoreGasUsed(tx) {
			break
		}
		g, _ := strconv.ParseUint(tx.GasUsed, 10, 64)
		sum += g
	}
	return sum
}

func blockGasFromEvents(events []abciEvent) uint64 {
	for _, ev := range events {
		if ev.Type != "block_gas" {
			continue
		}
		if amt := attrValue(ev.Attributes, "amount"); amt != "" {
			g, _ := strconv.ParseUint(amt, 10, 64)
			return g
		}
	}
	return 0
}

func baseFeeFromBeginEvents(events []abciEvent) string {
	for _, ev := range events {
		if ev.Type != "fee_market" && ev.Type != "feemarket" {
			continue
		}
		if bf := attrValue(ev.Attributes, "base_fee"); bf != "" {
			return bf
		}
	}
	return ""
}

// ParseBlockResults parses a block_results JSON response.
func ParseBlockResults(raw blockResultsResp) BlockResultsSummary {
	s := BlockResultsSummary{OK: true}
	s.GasUsedSum = sumBlockGasUsed(raw.Result.TxsResults)
	s.BlockGasWanted = blockGasFromEvents(raw.Result.EndBlockEvents)
	s.BaseFeeEvent = baseFeeFromBeginEvents(raw.Result.BeginBlockEvents)
	return s
}

// FetchBlockResults loads block_results for a given height (nil = latest).
func FetchBlockResults(rpc string, height int64) BlockResultsSummary {
	var raw blockResultsResp
	url := rpc + "/block_results"
	if height > 0 {
		url = fmt.Sprintf("%s/block_results?height=%d", rpc, height)
	}
	if err := doJSON(url, &raw); err != nil {
		return BlockResultsSummary{}
	}
	return ParseBlockResults(raw)
}
