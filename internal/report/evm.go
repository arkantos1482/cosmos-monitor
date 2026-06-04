package report

import "strings"

const (
	DefaultJSONRPCAPIs       = "eth,txpool,net,debug,web3"
	DefaultTxpoolGlobalSlots = 5120
	DefaultTxpoolGlobalQueue = 1024
)

// EVMWSEndpoint derives a WebSocket URL from the HTTP JSON-RPC endpoint.
func EVMWSEndpoint(httpURL string) string {
	u := strings.Replace(httpURL, "https://", "wss://", 1)
	u = strings.Replace(u, "http://", "ws://", 1)
	if strings.Contains(u, ":8545") {
		return strings.Replace(u, ":8545", ":8546", 1)
	}
	if strings.HasSuffix(u, "/") {
		return strings.TrimSuffix(u, "/") + ":8546"
	}
	return u + "  _(WS usually :8546)_"
}
