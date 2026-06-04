package main

import (
	"fmt"
	"html"
	"io"
	"sort"
	"strconv"
	"strings"
)

// Defaults aligned with tools/ops/deploy/configs/app.toml (EVM mempool + json-rpc).
const (
	defaultJSONRPCAPIs     = "eth,txpool,net,debug,web3"
	defaultTxpoolGlobalSlots = 5120
	defaultTxpoolGlobalQueue = 1024
)

func evmWSEndpoint(httpURL string) string {
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

func evmDisplaySymbol(denom string) string {
	switch strings.ToLower(denom) {
	case "apmt", "upmt":
		return "PMT"
	default:
		if denom == "" {
			return "PMT"
		}
		return strings.ToUpper(denom)
	}
}

func evmRPCOverallStatus(d WebData) (label, cssClass string) {
	if !d.EVMRPCOk {
		return "DOWN", "evm-pill-err"
	}
	if d.EVMBlockAgeErr || !d.EVMSynced || d.RPCProbeOK < d.RPCProbeTotal {
		return "DEGRADED", "evm-pill-warn"
	}
	return "OK", "evm-pill-ok"
}

func probeNamespace(method string) string {
	if i := strings.IndexByte(method, '_'); i > 0 {
		return method[:i]
	}
	return "other"
}

func sortedRPCProbes(probes []WebRPCProbe) []WebRPCProbe {
	out := append([]WebRPCProbe(nil), probes...)
	sort.SliceStable(out, func(i, j int) bool {
		pi, pj := out[i], out[j]
		ni, nj := probeNamespace(pi.Method), probeNamespace(pj.Method)
		if ni != nj {
			return ni < nj
		}
		if pi.OK != pj.OK {
			return !pi.OK // failures first within namespace
		}
		return pi.Method < pj.Method
	})
	return out
}

func probeLatencyClass(latency string) string {
	ms := probeLatencyMs(latency)
	switch {
	case ms < 0:
		return ""
	case ms >= 500:
		return "evm-lat-slow"
	case ms >= 200:
		return "evm-lat-warn"
	default:
		return "evm-lat-ok"
	}
}

func probeLatencyMs(latency string) int {
	latency = strings.TrimSpace(strings.TrimSuffix(latency, "ms"))
	v, err := strconv.Atoi(latency)
	if err != nil {
		return -1
	}
	return v
}

func jsonRPCCurl(endpoint, requestJSON string) string {
	escaped := strings.ReplaceAll(requestJSON, `'`, `'\''`)
	return fmt.Sprintf("curl -sS -X POST %s \\\n  -H 'Content-Type: application/json' \\\n  -d '%s'", endpoint, escaped)
}

func writeEVMRPCSection(w io.Writer, d WebData, web bool) {
	hint := func(text string) { fmt.Fprintf(w, "_%s_\n\n", text) }
	subsection := func(name string) { fmt.Fprintf(w, "\n## %s\n\n", name) }
	row := func(label, value string) { fmt.Fprintf(w, "- **%s**: %s\n", label, value) }

	overall, overallClass := evmRPCOverallStatus(d)

	syncLabel := "synced"
	if !d.EVMSynced {
		syncLabel = "syncing"
	}

	blockAge := "—"
	if d.EVMBlockAge != "" {
		blockAge = d.EVMBlockAge
		if d.EVMBlockAgeErr {
			blockAge += " ⚠ stalled"
		} else if d.EVMBlockAgeWarn {
			blockAge += " ⚠ slow"
		}
	}

	probeSummary := fmt.Sprintf("%d/%d probes", d.RPCProbeOK, d.RPCProbeTotal)
	listenLabel := "not listening"
	listenClass := "evm-pill-warn"
	if d.EVMListening {
		listenLabel = "listening"
		listenClass = "evm-pill-ok"
	}

	if web {
		fmt.Fprint(w, `<div class="evm-rpc-strip">`+"\n")
		writeEVMPill(w, overall, overallClass)
		writeEVMPill(w, "block "+html.EscapeString(blockAge), pillClassForBlockAge(d))
		writeEVMPill(w, syncLabel, pillClassForSync(d.EVMSynced))
		writeEVMPill(w, probeSummary, pillClassForProbes(d))
		writeEVMPill(w, listenLabel, listenClass)
		fmt.Fprint(w, `</div>`+"\n\n")
	} else {
		fmt.Fprintf(w, "**%s** · block %s · %s · %s · %s\n\n",
			overall, blockAge, syncLabel, probeSummary, listenLabel)
	}

	subsection("For operators")
	hint("HTTP/WS bind addresses from node `app.toml` `[json-rpc]`; APIs list is the deployed default for PMT.")
	httpEP := d.EVMHTTPEndpoint
	if httpEP == "" {
		httpEP = "http://localhost:8545"
	}
	wsEP := d.EVMWSEndpoint
	if wsEP == "" {
		wsEP = evmWSEndpoint(httpEP)
	}
	apis := d.JSONRPCAPIs
	if apis == "" {
		apis = defaultJSONRPCAPIs
	}
	row("HTTP JSON-RPC", "`"+httpEP+"`")
	row("WebSocket", "`"+wsEP+"`")
	row("enabled APIs", "`"+apis+"`")
	row("chain ID", fmt.Sprintf("%d  _(eth_chainId · MetaMask custom network)_", d.EVMChainID))
	if d.EVMDenom != "" {
		row("native denom", "`"+d.EVMDenom+"`")
	}
	if d.EVMClient != "" {
		row("client", d.EVMClient+"  _(web3_clientVersion)_")
	}

	symbol := evmDisplaySymbol(d.EVMDenom)
	networkName := strings.ToUpper(d.Network)
	if networkName == "" {
		networkName = "PMT"
	}
	wallet := fmt.Sprintf("Network name: %s\nRPC URL: %s\nChain ID: %d\nCurrency symbol: %s",
		networkName, httpEP, d.EVMChainID, symbol)
	if web {
		fmt.Fprintf(w, "\n<div class=\"evm-wallet-snippet\"><pre>%s</pre></div>\n\n",
			html.EscapeString(wallet))
	} else {
		fmt.Fprintf(w, "\n```text\n%s\n```\n\n", wallet)
	}

	subsection("Live (JSON-RPC)")
	hint("`eth_*` / `txpool_*` probes on each refresh; gas price also feeds §4 fee market.")
	row("block height", d.EVMBlock+"  _(eth_blockNumber)_")
	if d.EVMBlockAge != "" {
		ageStr := d.EVMBlockAge + "  _(eth_getBlockByNumber timestamp)_"
		if d.EVMBlockAgeErr {
			ageStr += "  ⚠ stalled"
		} else if d.EVMBlockAgeWarn {
			ageStr += "  ⚠ slow"
		}
		row("last block age", ageStr)
	}
	row("sync", syncLabel+"  _(eth_syncing)_")
	if d.GasPrice != "" {
		row("gas price", d.GasPrice+"  _(eth_gasPrice)_")
	}
	txpool := fmt.Sprintf("pending %s · queued %s  _(txpool_status)_",
		formatTxpoolCount(d.PendingTx, d.TxpoolGlobalSlots),
		formatTxpoolCount(d.QueuedTx, d.TxpoolGlobalQueue))
	row("txpool", txpool)
	row("EVM peers", fmt.Sprintf("%d  _(net_peerCount — often 0 on validators)_", d.EVMPeerCount))

	subsection("Probe health")
	hint("Client-side `POST` JSON-RPC 2.0; latency measured on this host. Failed methods expand with curl + bodies.")
	if web {
		writeEVMProbeTableWeb(w, d, httpEP)
	} else {
		writeEVMProbeTableTerm(w, d, httpEP)
	}
}

func formatTxpoolCount(n, limit uint64) string {
	if limit == 0 {
		return fmt.Sprintf("%d", n)
	}
	return fmt.Sprintf("%d / %d", n, limit)
}

func pillClassForBlockAge(d WebData) string {
	switch {
	case d.EVMBlockAgeErr:
		return "evm-pill-err"
	case d.EVMBlockAgeWarn:
		return "evm-pill-warn"
	default:
		return ""
	}
}

func pillClassForSync(synced bool) string {
	if synced {
		return "evm-pill-ok"
	}
	return "evm-pill-warn"
}

func pillClassForProbes(d WebData) string {
	if d.RPCProbeOK < d.RPCProbeTotal {
		return "evm-pill-warn"
	}
	return "evm-pill-ok"
}

func writeEVMPill(w io.Writer, label, extraClass string) {
	class := "evm-pill"
	if extraClass != "" {
		class += " " + extraClass
	}
	fmt.Fprintf(w, `<span class="%s">%s</span>`, class, html.EscapeString(label))
}

func writeEVMProbeTableWeb(w io.Writer, d WebData, endpoint string) {
	probes := sortedRPCProbes(d.RPCProbes)
	lastNS := ""
	fmt.Fprint(w, `<table class="evm-probe-table"><thead><tr>`+
		`<th>method</th><th>status</th><th>latency</th><th>error</th></tr></thead><tbody>`+"\n")
	for _, p := range probes {
		ns := probeNamespace(p.Method)
		if ns != lastNS {
			fmt.Fprintf(w, `<tr class="evm-probe-group"><td colspan="4">%s</td></tr>`+"\n", html.EscapeString(ns))
			lastNS = ns
		}
		statusClass := "evm-probe-ok"
		status := "ok"
		if !p.OK {
			statusClass = "evm-probe-fail"
			status = "FAIL"
		}
		latClass := probeLatencyClass(p.Latency)
		errCell := html.EscapeString(truncate(p.Error, 48))
		fmt.Fprintf(w, `<tr class="%s %s"><td><code>%s</code></td><td>%s</td><td class="%s">%s</td><td>%s</td></tr>`+"\n",
			statusClass, latClass, html.EscapeString(p.Method), status, latClass, html.EscapeString(p.Latency), errCell)
	}
	fmt.Fprint(w, `</tbody></table>`+"\n\n")

	for _, p := range probes {
		if p.OK {
			continue
		}
		fmt.Fprintf(w, `<details class="evm-probe-detail" open><summary><code>%s</code> FAIL · %s</summary>`+"\n",
			html.EscapeString(p.Method), html.EscapeString(p.Latency))
		if p.Error != "" {
			fmt.Fprintf(w, `<p class="evm-probe-err">%s</p>`+"\n", html.EscapeString(p.Error))
		}
		fmt.Fprintf(w, `<pre class="evm-curl">%s</pre>`+"\n", html.EscapeString(jsonRPCCurl(endpoint, p.Request)))
		fmt.Fprintf(w, `<pre class="evm-json">%s`+"\n"+`→`+"\n"+`%s</pre>`+"\n",
			html.EscapeString(p.Request), html.EscapeString(p.Response))
		fmt.Fprint(w, `</details>`+"\n\n")
	}
}

func writeEVMProbeTableTerm(w io.Writer, d WebData, endpoint string) {
	probes := sortedRPCProbes(d.RPCProbes)
	lastNS := ""
	fmt.Fprintf(w, "| method | status | latency | error |\n|--------|--------|---------|-------|\n")
	for _, p := range probes {
		ns := probeNamespace(p.Method)
		if ns != lastNS {
			fmt.Fprintf(w, "| _%s_ | | | |\n", ns)
			lastNS = ns
		}
		status := "ok"
		errStr := ""
		if !p.OK {
			status = "FAIL"
			errStr = truncate(p.Error, 40)
		}
		fmt.Fprintf(w, "| `%s` | %s | %s | %s |\n", p.Method, status, p.Latency, errStr)
	}
	fmt.Fprintln(w)

	for _, p := range probes {
		if p.OK {
			continue
		}
		fmt.Fprintf(w, "**%s** FAIL (%s)\n\n", p.Method, p.Latency)
		if p.Error != "" {
			fmt.Fprintf(w, "error: %s\n\n", p.Error)
		}
		fmt.Fprintf(w, "```bash\n%s\n```\n\n", jsonRPCCurl(endpoint, p.Request))
		fmt.Fprintf(w, "```json\n%s\n→\n%s\n```\n\n", p.Request, p.Response)
	}
}
