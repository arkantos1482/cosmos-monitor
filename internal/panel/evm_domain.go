package panel

import (
	"fmt"
	"html"
	"sort"
	"strings"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
	"github.com/arkantos1482/cosmos-monitor/internal/report"
)

func evmRPCHealthCardsHTML(d model.Report) string {
	return ecoDomainsWrap(
		evmReachabilityCardHTML(d),
		evmChainHeadCardHTML(d),
		evmTxpoolCardHTML(d),
		evmNetCardHTML(d),
	)
}

func evmReachabilityCardHTML(d model.Report) string {
	var b strings.Builder
	status := "DOWN"
	badge := "bad"
	if d.EVMRPCOk {
		status = "UP"
		badge = "ok"
		if d.RPCProbeOK < d.RPCProbeTotal {
			status = "DEGRADED"
			badge = "warn"
		}
	}
	fmt.Fprintf(&b, `<div class="eco-domain eco-domain--rpc-reach">`)
	ecoDomainCardTitle(&b, "Reachability", "JSON-RPC transport", badge, status)
	b.WriteString(`<div class="eco-domain__rows">`)

	listen := "not listening"
	if d.EVMListening {
		listen = "listening"
	}
	ecoDomainRow(&b, "", "HTTP endpoint", orEcoDash(d.EVMHTTPEndpoint), "POST target for probes")
	ecoDomainRow(&b, "", "net_listening", listen, "socket accepting connections")
	ecoDomainRow(&b, "", "probes", fmt.Sprintf("%d / %d ok", d.RPCProbeOK, d.RPCProbeTotal), "method health checks this refresh")

	ecoDomainCardClose(&b)
	return b.String()
}

func evmChainHeadCardHTML(d model.Report) string {
	var b strings.Builder
	syncLabel := "synced"
	badge, badgeClass := "ok", "ok"
	if !d.EVMSynced {
		syncLabel = "syncing"
		badge, badgeClass = "SYNC", "warn"
	}
	if d.EVMBlockAgeErr {
		badge, badgeClass = "STALE", "bad"
	} else if d.EVMBlockAgeWarn {
		badge, badgeClass = "SLOW", "warn"
	}
	fmt.Fprintf(&b, `<div class="eco-domain eco-domain--rpc-head">`)
	ecoDomainCardTitle(&b, "Chain head", "eth_* block probes", badgeClass, badge)
	b.WriteString(`<div class="eco-domain__rows">`)

	ecoDomainRow(&b, "", "eth_blockNumber", orEcoDash(d.EVMBlock), "latest block height")
	if d.EVMBlockAge != "" {
		age := d.EVMBlockAge
		if d.EVMBlockAgeErr {
			age += " (stalled)"
		} else if d.EVMBlockAgeWarn {
			age += " (slow)"
		}
		ecoDomainRow(&b, "", "block age", age, "eth_getBlockByNumber timestamp")
	}
	ecoDomainRow(&b, "", "eth_syncing", syncLabel, "false when caught up")

	ecoDomainCardClose(&b)
	return b.String()
}

func evmTxpoolCardHTML(d model.Report) string {
	var b strings.Builder
	ecoDomainCardOpen(&b, "eco-domain--rpc-txpool", "Txpool", "txpool_* probes")

	pending := formatTxpoolCount(d.PendingTx, d.TxpoolGlobalSlots)
	queued := formatTxpoolCount(d.QueuedTx, d.TxpoolGlobalQueue)
	ecoDomainRow(&b, "", "pending", pending, "txpool_status.pending")
	ecoDomainRow(&b, "", "queued", queued, "txpool_status.queued")

	ecoDomainCardClose(&b)
	return b.String()
}

func evmNetCardHTML(d model.Report) string {
	var b strings.Builder
	ecoDomainCardOpen(&b, "eco-domain--rpc-net", "Network", "net_* / web3_* probes")

	if d.EVMChainID > 0 {
		ecoDomainRow(&b, "", "eth_chainId", fmt.Sprintf("%d", d.EVMChainID), "wallet network ID")
	}
	if d.EVMClient != "" {
		ecoDomainRow(&b, "", "web3_clientVersion", d.EVMClient, "EVM client build")
	}
	ecoDomainRow(&b, "", "net_peerCount", fmt.Sprintf("%d", d.EVMPeerCount), "execution-layer peers (often 0 on validators)")

	ecoDomainCardClose(&b)
	return b.String()
}

func evmWalletCardHTML(d model.Report) string {
	wsEP := d.EVMWSEndpoint
	if wsEP == "" {
		httpEP := d.EVMHTTPEndpoint
		if httpEP == "" {
			httpEP = "http://localhost:8545"
		}
		wsEP = report.EVMWSEndpoint(httpEP)
	}
	apis := d.JSONRPCAPIs
	if apis == "" {
		apis = report.DefaultJSONRPCAPIs
	}

	var b strings.Builder
	ecoDomainCardOpen(&b, "eco-domain--rpc-wallet", "Wallet", "dApp / MetaMask")
	ecoDomainRow(&b, "", "WebSocket", wsEP, "subscriptions and event filters")
	ecoDomainRow(&b, "", "enabled APIs", apis, "namespaces exposed by this node")
	ecoDomainCardClose(&b)
	return b.String()
}

func evmRPCProbeTableHTML(d model.Report) string {
	probes := sortedRPCProbes(d.RPCProbes)
	if len(probes) == 0 {
		return `<p class="evm-probes__empty">No JSON-RPC probes recorded.</p>`
	}

	var b strings.Builder
	b.WriteString(`<table class="evm-probes__table"><thead><tr>`)
	b.WriteString(`<th class="evm-probes__mark-hdr" aria-hidden="true"></th>`)
	b.WriteString(`<th class="evm-probes__col-method">method</th>`)
	b.WriteString(`<th class="evm-probes__col-checks">checks</th>`)
	b.WriteString(`<th class="evm-probes__col-status">status</th>`)
	b.WriteString(`<th class="evm-probes__col-latency">latency</th>`)
	b.WriteString(`</tr></thead><tbody>`)
	for _, p := range probes {
		mark := "·"
		rowClass := ""
		status := "ok"
		if !p.OK {
			mark = "✗"
			rowClass = ` class="dash-sources__row--fail"`
			status = "FAIL"
		}
		checks := rpcProbeHint(p.Method)
		if !p.OK && p.Error != "" {
			checks = p.Error
		}
		fmt.Fprintf(&b, `<tr%s><td class="evm-probes__mark">%s</td>`, rowClass, mark)
		fmt.Fprintf(&b, `<td class="evm-probes__method"><span class="evm-probes__ns">%s</span>`,
			html.EscapeString(probeNamespace(p.Method)))
		fmt.Fprintf(&b, `<span class="evm-probes__name">%s</span></td>`, html.EscapeString(p.Method))
		fmt.Fprintf(&b, `<td class="evm-probes__checks">%s</td>`, html.EscapeString(checks))
		fmt.Fprintf(&b, `<td class="evm-probes__status">%s</td>`, html.EscapeString(status))
		fmt.Fprintf(&b, `<td class="evm-probes__latency">%s</td></tr>`, html.EscapeString(p.Latency))
	}
	b.WriteString(`</tbody></table>`)
	return b.String()
}

func rpcProbeHint(method string) string {
	switch method {
	case "eth_blockNumber":
		return "latest block height"
	case "eth_chainId":
		return "network ID for wallets"
	case "eth_syncing":
		return "sync status (false = caught up)"
	case "eth_getBlockByNumber":
		return "latest block header + timestamp"
	case "txpool_status":
		return "mempool pending / queued counts"
	case "net_peerCount":
		return "execution-layer peer count"
	case "net_listening":
		return "RPC socket accepting connections"
	case "web3_clientVersion":
		return "client identity string"
	default:
		return "JSON-RPC probe"
	}
}

func probeNamespace(method string) string {
	if i := strings.IndexByte(method, '_'); i > 0 {
		return method[:i]
	}
	return "other"
}

func sortedRPCProbes(probes []model.RPCProbe) []model.RPCProbe {
	out := append([]model.RPCProbe(nil), probes...)
	sort.SliceStable(out, func(i, j int) bool {
		pi, pj := out[i], out[j]
		ni, nj := probeNamespace(pi.Method), probeNamespace(pj.Method)
		if pi.OK != pj.OK {
			return !pi.OK
		}
		if ni != nj {
			return ni < nj
		}
		return pi.Method < pj.Method
	})
	return out
}

func evmProbeNamespaces(probes []model.RPCProbe) []string {
	seen := map[string]bool{}
	var out []string
	for _, p := range sortedRPCProbes(probes) {
		ns := probeNamespace(p.Method)
		if seen[ns] {
			continue
		}
		seen[ns] = true
		out = append(out, ns)
	}
	return out
}

func evmNamespaceOK(probes []model.RPCProbe, ns string) bool {
	found := false
	for _, p := range probes {
		if probeNamespace(p.Method) != ns {
			continue
		}
		found = true
		if !p.OK {
			return false
		}
	}
	return found
}
