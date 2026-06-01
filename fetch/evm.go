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
	BlockNumber        uint64
	ChainID            uint64
	Syncing            bool
	GasPrice           string
	PendingTx          uint64
	QueuedTx           uint64
	PeerCount          uint64
	ClientVersion      string
	NetListening       bool
	EVMBlockTimestamp  uint64 // unix seconds of latest EVM block
	Err                error
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
	return evmCallP(endpoint, method, []any{}, target)
}

func evmCallP(endpoint, method string, params []any, target any) error {
	body, _ := json.Marshal(rpcRequest{JSONRPC: "2.0", Method: method, Params: params, ID: 1})
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
		gp := hexToUint64(gasPriceHex)
		if gp == 0 {
			snap.GasPrice = "0"
		} else {
			snap.GasPrice = FormatCoin(fmt.Sprintf("%d", gp), "apmt")
		}
	}

	var txpool struct {
		Pending string `json:"pending"`
		Queued  string `json:"queued"`
	}
	if err := evmCall(endpoint, "txpool_status", &txpool); err == nil {
		snap.PendingTx = hexToUint64(txpool.Pending)
		snap.QueuedTx = hexToUint64(txpool.Queued)
	}

	// net_peerCount returns a plain integer on this chain, not a hex string.
	var peerRaw json.RawMessage
	if err := evmCall(endpoint, "net_peerCount", &peerRaw); err == nil {
		s := strings.Trim(string(peerRaw), `"`)
		if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
			snap.PeerCount = hexToUint64(s)
		} else {
			v, _ := strconv.ParseUint(s, 10, 64)
			snap.PeerCount = v
		}
	}

	var clientVer string
	if err := evmCall(endpoint, "web3_clientVersion", &clientVer); err == nil {
		snap.ClientVersion = clientVer
	}

	var listening bool
	if err := evmCall(endpoint, "net_listening", &listening); err == nil {
		snap.NetListening = listening
	}

	var latestBlock struct {
		Timestamp string `json:"timestamp"`
	}
	if err := evmCallP(endpoint, "eth_getBlockByNumber", []any{"latest", false}, &latestBlock); err == nil {
		snap.EVMBlockTimestamp = hexToUint64(latestBlock.Timestamp)
	}

	return snap
}
