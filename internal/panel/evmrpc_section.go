package panel

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
	"github.com/arkantos1482/cosmos-monitor/internal/report"
)

const maxProbeJSONBytes = 12_000

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

func evmRPCOverallStatus(d model.Report) string {
	if !d.EVMRPCOk {
		return "DOWN"
	}
	if d.EVMBlockAgeErr || !d.EVMSynced || d.RPCProbeOK < d.RPCProbeTotal {
		return "DEGRADED"
	}
	return "OK"
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
		if ni != nj {
			return ni < nj
		}
		if pi.OK != pj.OK {
			return !pi.OK
		}
		return pi.Method < pj.Method
	})
	return out
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

func writeEVMRPCSection(w Writer, d model.Report) {
	overall := evmRPCOverallStatus(d)

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
	if d.EVMListening {
		listenLabel = "listening"
	}

	w.StrongLine(fmt.Sprintf("**RPC: %s** · block %s · %s · %s · %s",
		overall, blockAge, syncLabel, probeSummary, listenLabel))

	w.Subsection("For operators")
	w.Hint("HTTP/WS bind addresses from node `app.toml` `[json-rpc]`; APIs list is the deployed default for PMT.")
	httpEP := d.EVMHTTPEndpoint
	if httpEP == "" {
		httpEP = "http://localhost:8545"
	}
	wsEP := d.EVMWSEndpoint
	if wsEP == "" {
		wsEP = report.EVMWSEndpoint(httpEP)
	}
	apis := d.JSONRPCAPIs
	if apis == "" {
		apis = report.DefaultJSONRPCAPIs
	}
	w.Row("HTTP JSON-RPC", "`"+httpEP+"`")
	w.Row("WebSocket", "`"+wsEP+"`")
	w.Row("enabled APIs", "`"+apis+"`")
	w.Row("chain ID", fmt.Sprintf("%d  _(eth_chainId · MetaMask custom network)_", d.EVMChainID))
	if d.EVMDenom != "" {
		w.Row("native denom", "`"+d.EVMDenom+"`")
	}
	if d.EVMClient != "" {
		w.Row("client", d.EVMClient+"  _(web3_clientVersion)_")
	}

	symbol := evmDisplaySymbol(d.EVMDenom)
	networkName := strings.ToUpper(d.Network)
	if networkName == "" {
		networkName = "PMT"
	}
	wallet := fmt.Sprintf("Network name: %s\nRPC URL: %s\nChain ID: %d\nCurrency symbol: %s",
		networkName, httpEP, d.EVMChainID, symbol)
	w.Pre(wallet)

	w.Subsection("Live (JSON-RPC)")
	w.Hint("`eth_*` / `txpool_*` probes on each refresh; gas price also feeds §4 fee market.")
	w.Row("block height", d.EVMBlock+"  _(eth_blockNumber)_")
	if d.EVMBlockAge != "" {
		ageStr := d.EVMBlockAge + "  _(eth_getBlockByNumber timestamp)_"
		if d.EVMBlockAgeErr {
			ageStr += "  ⚠ stalled"
		} else if d.EVMBlockAgeWarn {
			ageStr += "  ⚠ slow"
		}
		w.Row("last block age", ageStr)
	}
	w.Row("sync", syncLabel+"  _(eth_syncing)_")
	if d.GasPrice != "" {
		w.Row("gas price", d.GasPrice+"  _(eth_gasPrice)_")
	}
	txpool := fmt.Sprintf("pending %s · queued %s  _(txpool_status)_",
		formatTxpoolCount(d.PendingTx, d.TxpoolGlobalSlots),
		formatTxpoolCount(d.QueuedTx, d.TxpoolGlobalQueue))
	w.Row("txpool", txpool)
	w.Row("EVM peers", fmt.Sprintf("%d  _(net_peerCount — often 0 on validators)_", d.EVMPeerCount))

	w.Subsection("Probe health")
	w.Hint("Client-side `POST` JSON-RPC 2.0; each method shows the request line and pretty-printed response body from the last refresh.")
	writeEVMProbeLog(w, d, httpEP)
}

func formatTxpoolCount(n, limit uint64) string {
	if limit == 0 {
		return fmt.Sprintf("%d", n)
	}
	return fmt.Sprintf("%d / %d", n, limit)
}

func renderProbeLog(probes []model.RPCProbe) string {
	const (
		padMethod  = 24
		padStatus  = 6
		padLatency = 7
	)
	var b strings.Builder
	b.WriteString("  method                    status  latency\n")
	b.WriteString("  ─────────────────────────────────────────\n")

	lastNS := ""
	for _, p := range sortedRPCProbes(probes) {
		ns := probeNamespace(p.Method)
		if ns != lastNS {
			if lastNS != "" {
				b.WriteByte('\n')
			}
			fmt.Fprintf(&b, "  [%s]\n", strings.ToUpper(ns))
			lastNS = ns
		}
		status := "ok"
		mark := "·"
		if !p.OK {
			status = "FAIL"
			mark = "✗"
		}
		line := fmt.Sprintf("  %s  %-*s  %-*s  %-*s",
			mark, padMethod, p.Method, padStatus, status, padLatency, p.Latency)
		if !p.OK && p.Error != "" {
			line += "  " + report.Truncate(p.Error, 44)
		}
		b.WriteString(line + "\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

func probeStatusLabel(ok bool) string {
	if ok {
		return "ok"
	}
	return "FAIL"
}

func prettyProbeJSON(raw string, maxBytes int) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "(empty)"
	}
	var v any
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		return truncateJSON(raw, maxBytes)
	}
	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return truncateJSON(raw, maxBytes)
	}
	return truncateJSON(string(out), maxBytes)
}

func truncateJSON(s string, maxBytes int) string {
	if maxBytes <= 0 || len(s) <= maxBytes {
		return s
	}
	return s[:maxBytes] + "\n… (truncated)"
}

func formatProbeExchange(p model.RPCProbe) string {
	var b strings.Builder
	fmt.Fprintf(&b, "── %s · %s · %s ──\n", p.Method, probeStatusLabel(p.OK), p.Latency)
	if !p.OK && p.Error != "" {
		fmt.Fprintf(&b, "err » %s\n", p.Error)
	}
	if p.Request != "" {
		fmt.Fprintf(&b, "req » %s\n", strings.TrimSpace(p.Request))
	}
	b.WriteString("res » ")
	res := prettyProbeJSON(p.Response, maxProbeJSONBytes)
	if !strings.Contains(res, "\n") {
		b.WriteString(res + "\n")
		return strings.TrimRight(b.String(), "\n")
	}
	lines := strings.Split(res, "\n")
	b.WriteString(lines[0] + "\n")
	for _, line := range lines[1:] {
		b.WriteString("      " + line + "\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

func writeEVMProbeLog(w Writer, d model.Report, endpoint string) {
	log := renderProbeLog(d.RPCProbes)
	w.Pre(log)

	for _, p := range sortedRPCProbes(d.RPCProbes) {
		body := formatProbeExchange(p)
		w.Pre(body)
		if !p.OK {
			w.PreBash(jsonRPCCurl(endpoint, p.Request))
		}
	}
}
