package fetch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"
)

// RPCProbe holds one JSON-RPC health check with raw request/response.
type RPCProbe struct {
	Method   string
	Request  string
	Response string
	OK       bool
	Error    string
	Latency  time.Duration
}

// EVMSnapshot holds EVM JSON-RPC data.
type EVMSnapshot struct {
	BlockNumber       uint64
	ChainID           uint64
	Syncing           bool
	GasPrice          string
	PendingTx         uint64
	QueuedTx          uint64
	PeerCount         uint64
	ClientVersion     string
	NetListening      bool
	EVMBlockTimestamp uint64 // unix seconds of latest EVM block
	Probes            []RPCProbe
	Err               error
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

func evmProbe(endpoint, method string, params []any) RPCProbe {
	req := rpcRequest{JSONRPC: "2.0", Method: method, Params: params, ID: 1}
	reqBody, _ := json.Marshal(req)
	p := RPCProbe{
		Method:  method,
		Request: compactJSONFull(reqBody),
	}

	start := time.Now()
	resp, err := httpClient.Post(endpoint, "application/json", bytes.NewReader(reqBody))
	p.Latency = time.Since(start)
	if err != nil {
		p.Error = err.Error()
		return p
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		p.Error = err.Error()
		return p
	}
	p.Response = compactJSONFull(raw)
	if resp.StatusCode != 200 {
		p.Error = fmt.Sprintf("HTTP %d", resp.StatusCode)
		return p
	}

	var rpc rpcResponse
	if err := json.Unmarshal(raw, &rpc); err != nil {
		p.Error = err.Error()
		return p
	}
	if rpc.Error != nil {
		p.Error = rpc.Error.Message
		return p
	}
	p.OK = true
	return p
}

func compactJSONFull(b []byte) string {
	var buf bytes.Buffer
	if err := json.Compact(&buf, b); err != nil {
		return string(b)
	}
	return buf.String()
}

func TruncateJSON(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "…"
}

func hexToUint64(h string) uint64 {
	h = strings.TrimPrefix(h, "0x")
	v, _ := strconv.ParseUint(h, 16, 64)
	return v
}

func probeResult[T any](p RPCProbe, target *T) bool {
	if !p.OK {
		return false
	}
	var rpc rpcResponse
	if err := json.Unmarshal([]byte(p.Response), &rpc); err != nil {
		return false
	}
	return json.Unmarshal(rpc.Result, target) == nil
}

// FetchEVM fetches all EVM JSON-RPC metrics and probe traces.
func FetchEVM(endpoint string) EVMSnapshot {
	snap := EVMSnapshot{}

	methods := []struct {
		method string
		params []any
	}{
		{"eth_blockNumber", nil},
		{"eth_chainId", nil},
		{"eth_syncing", nil},
		{"eth_gasPrice", nil},
		{"txpool_status", nil},
		{"net_peerCount", nil},
		{"net_listening", nil},
		{"web3_clientVersion", nil},
		{"eth_getBlockByNumber", []any{"latest", false}},
	}

	probes := make([]RPCProbe, len(methods))
	var probeWG sync.WaitGroup
	for i, m := range methods {
		i, m := i, m
		probeWG.Add(1)
		go func() {
			defer probeWG.Done()
			params := m.params
			if params == nil {
				params = []any{}
			}
			probes[i] = evmProbe(endpoint, m.method, params)
		}()
	}
	probeWG.Wait()
	snap.Probes = probes

	byMethod := map[string]RPCProbe{}
	for _, p := range snap.Probes {
		byMethod[p.Method] = p
	}

	blockProbe := byMethod["eth_blockNumber"]
	if !blockProbe.OK {
		snap.Err = fmt.Errorf("eth_blockNumber: %s", firstNonEmpty(blockProbe.Error, "failed"))
		return snap
	}
	var blockHex string
	if probeResult(blockProbe, &blockHex) {
		snap.BlockNumber = hexToUint64(blockHex)
	}

	var chainIDHex string
	if probeResult(byMethod["eth_chainId"], &chainIDHex) {
		snap.ChainID = hexToUint64(chainIDHex)
	}

	var syncRaw json.RawMessage
	if probeResult(byMethod["eth_syncing"], &syncRaw) {
		snap.Syncing = string(syncRaw) != "false"
	}

	var gasPriceHex string
	if probeResult(byMethod["eth_gasPrice"], &gasPriceHex) {
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
	if probeResult(byMethod["txpool_status"], &txpool) {
		snap.PendingTx = hexToUint64(txpool.Pending)
		snap.QueuedTx = hexToUint64(txpool.Queued)
	}

	var peerRaw json.RawMessage
	if probeResult(byMethod["net_peerCount"], &peerRaw) {
		s := strings.Trim(string(peerRaw), `"`)
		if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
			snap.PeerCount = hexToUint64(s)
		} else {
			v, _ := strconv.ParseUint(s, 10, 64)
			snap.PeerCount = v
		}
	}

	var clientVer string
	if probeResult(byMethod["web3_clientVersion"], &clientVer) {
		if i := strings.IndexByte(clientVer, '\n'); i >= 0 {
			clientVer = clientVer[:i]
		}
		snap.ClientVersion = strings.TrimSpace(clientVer)
	}

	var listening bool
	if probeResult(byMethod["net_listening"], &listening) {
		snap.NetListening = listening
	}

	var latestBlock struct {
		Timestamp string `json:"timestamp"`
	}
	if probeResult(byMethod["eth_getBlockByNumber"], &latestBlock) {
		snap.EVMBlockTimestamp = hexToUint64(latestBlock.Timestamp)
	}

	return snap
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return "unknown"
}
