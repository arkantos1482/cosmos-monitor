package panel

import (
	"fmt"
	"html"
	"strings"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
	"github.com/arkantos1482/cosmos-monitor/internal/report"
)

func writeEVMSection(w Writer, d model.Report) {
	w.Section("3. EVM JSON-RPC")
	writeEmbeddedSectionIntro(w, "JSON-RPC health: live method probes, latency, block freshness, and wallet connection endpoints.")
	writeEVMSummary(w, d, SummaryEmbedded)
	writeEVMRPCSection(w, d)
	writeSectionSources(w, ViewEVM, d)
	w.BlankLine()
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

func writeEVMSummary(w Writer, d model.Report, mode SummaryMode) {
	overall := evmRPCOverallStatus(d)
	overallKind := "ok"
	switch overall {
	case "DOWN":
		overallKind = "bad"
	case "DEGRADED":
		overallKind = "warn"
	}
	syncLabel := "synced"
	if !d.EVMSynced {
		syncLabel = "syncing"
	}
	blockAge := "—"
	if d.EVMBlockAge != "" {
		blockAge = d.EVMBlockAge
	}
	listenLabel := "not listening"
	if d.EVMListening {
		listenLabel = "listening"
	}
	httpEP := d.EVMHTTPEndpoint
	if httpEP == "" {
		httpEP = "http://localhost:8545"
	}

	summaryWrapStart(w, mode, "evm")
	w.WriteHTML(`<div class="evm-summary">`)
	w.WriteHTML(`<div class="evm-summary__header">`)
	writeSummaryBadges(w, "evm-summary__status", summaryBadge{"RPC " + overall, overallKind})
	w.WriteHTML(fmt.Sprintf(`<span class="evm-summary__meta">block %s · %s · %s</span>`,
		html.EscapeString(blockAge), html.EscapeString(syncLabel), html.EscapeString(listenLabel)))
	w.WriteHTML(`</div>`)
	w.WriteHTML(fmt.Sprintf(`<p class="evm-summary__probes-label">Probes %d/%d ok</p>`, d.RPCProbeOK, d.RPCProbeTotal))
	w.WriteHTML(`<div class="evm-summary__probes">`)
	for _, ns := range evmProbeNamespaces(d.RPCProbes) {
		ok := evmNamespaceOK(d.RPCProbes, ns)
		cls := "evm-summary__probe--ok"
		if !ok {
			cls = "evm-summary__probe--fail"
		}
		w.WriteHTML(fmt.Sprintf(`<span class="evm-summary__probe %s" title="%s namespace">%s</span>`,
			cls, html.EscapeString(ns), html.EscapeString(ns)))
	}
	w.WriteHTML(`</div>`)
	w.WriteHTML(fmt.Sprintf(
		`<p class="evm-summary__detail">Chain <strong>%d</strong> · txpool <strong>%d</strong> pending · <strong>%d</strong> queued</p>`,
		d.EVMChainID, d.PendingTx, d.QueuedTx))
	w.WriteHTML(fmt.Sprintf(`<p class="evm-summary__endpoint mono">%s</p>`, html.EscapeString(report.Truncate(httpEP, 48))))
	w.WriteHTML(`</div>`)
	summaryWrapEnd(w, mode)
}

func writeEVMRPCSection(w Writer, d model.Report) {
	w.WriteHTML(evmRPCHealthCardsHTML(d))

	w.Subsection("Method probes")
	w.WriteHTML(evmRPCProbeTableHTML(d))

	w.Subsection("Wallet endpoints")
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

	symbol := evmDisplaySymbol(d.EVMDenom)
	networkName := strings.ToUpper(d.Network)
	if networkName == "" {
		networkName = "PMT"
	}
	wallet := fmt.Sprintf("Network name: %s\nRPC URL: %s\nChain ID: %d\nCurrency symbol: %s",
		networkName, httpEP, d.EVMChainID, symbol)
	w.Pre(wallet)
}

func formatTxpoolCount(n, limit uint64) string {
	if limit == 0 {
		return fmt.Sprintf("%d", n)
	}
	return fmt.Sprintf("%d / %d", n, limit)
}
