package fetch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// EVMSnapshot holds EVM JSON-RPC data.
type EVMSnapshot struct {
	BlockNumber uint64
	ChainID     uint64
	Syncing     bool
	GasPrice    string
	PendingTx   uint64
	QueuedTx    uint64
	PeerCount   uint64
	Err         error
}

type rpcRequest struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  []any  `json:"params"`
	ID      int    `json:"id"`
}

type rpcResponse struct {
	Result json.RawMessage `json:"result"`
	Error  *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func evmCall(endpoint, method string, target any) error {
	body, _ := json.Marshal(rpcRequest{JSONRPC: "2.0", Method: method, Params: []any{}, ID: 1})
	resp, err := httpClient.Post(endpoint, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var rpc rpcResponse
	if err := json.NewDecoder(resp.Body).Decode(&rpc); err != nil {
		return err
	}
	if rpc.Error != nil {
		return fmt.Errorf("rpc error: %s", rpc.Error.Message)
	}
	return json.Unmarshal(rpc.Result, target)
}

func hexToUint64(h string) uint64 {
	h = strings.TrimPrefix(h, "0x")
	v, _ := strconv.ParseUint(h, 16, 64)
	return v
}

// FetchEVM fetches all EVM JSON-RPC metrics.
func FetchEVM(endpoint string) EVMSnapshot {
	snap := EVMSnapshot{}

	var blockHex string
	if err := evmCall(endpoint, "eth_blockNumber", &blockHex); err != nil {
		snap.Err = fmt.Errorf("eth_blockNumber: %w", err)
		return snap
	}
	snap.BlockNumber = hexToUint64(blockHex)

	var chainIDHex string
	if err := evmCall(endpoint, "eth_chainId", &chainIDHex); err == nil {
		snap.ChainID = hexToUint64(chainIDHex)
	}

	// eth_syncing returns false (bool) or an object — handle both
	var syncRaw json.RawMessage
	if err := evmCall(endpoint, "eth_syncing", &syncRaw); err == nil {
		snap.Syncing = string(syncRaw) != "false"
	}

	var gasPriceHex string
	if err := evmCall(endpoint, "eth_gasPrice", &gasPriceHex); err == nil {
		snap.GasPrice = gasPriceHex
	}

	var txpool struct {
		Pending string `json:"pending"`
		Queued  string `json:"queued"`
	}
	if err := evmCall(endpoint, "txpool_status", &txpool); err == nil {
		snap.PendingTx = hexToUint64(txpool.Pending)
		snap.QueuedTx = hexToUint64(txpool.Queued)
	}

	var peerHex string
	if err := evmCall(endpoint, "net_peerCount", &peerHex); err == nil {
		snap.PeerCount = hexToUint64(peerHex)
	}

	return snap
}
