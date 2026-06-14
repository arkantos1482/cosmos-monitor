package panel

import (
	"fmt"
	"html"
	"strconv"
	"strings"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
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
	syncKind := "ok"
	if !d.EVMSynced {
		syncLabel = "syncing"
		syncKind = "warn"
	}
	listenLabel := "listening"
	listenKind := "ok"
	if !d.EVMListening {
		listenLabel = "not listening"
		listenKind = "warn"
	}
	blockAge, ageTone := evmBlockAgeKPI(d)
	httpEP := d.EVMHTTPEndpoint
	if httpEP == "" {
		httpEP = "http://localhost:8545"
	}

	probePct := 0
	if d.RPCProbeTotal > 0 {
		probePct = d.RPCProbeOK * 100 / d.RPCProbeTotal
	}
	heroClass := "evm-summary__hero-ok"
	switch {
	case !d.EVMRPCOk:
		heroClass += " evm-summary__hero-ok--bad"
	case probePct < 100:
		heroClass += " evm-summary__hero-ok--warn"
	}

	summaryWrapStart(w, mode, "evm")
	w.WriteHTML(`<div class="evm-summary">`)
	w.WriteHTML(`<div class="evm-summary__top">`)
	w.WriteHTML(`<div class="evm-summary__hero-wrap">`)
	w.WriteHTML(fmt.Sprintf(
		`<div class="evm-summary__hero"><span class="%s">%d</span><span class="evm-summary__hero-total">/%d</span></div>`,
		heroClass, d.RPCProbeOK, d.RPCProbeTotal))
	w.WriteHTML(`<p class="evm-summary__hero-label">probes passing</p>`)
	w.WriteHTML(`</div>`)
	writeSummaryBadges(w, "evm-summary__badges",
		summaryBadge{"RPC " + overall, overallKind},
		summaryBadge{listenLabel, listenKind},
		summaryBadge{syncLabel, syncKind},
	)
	w.WriteHTML(`</div>`)

	if d.RPCProbeTotal > 0 {
		writeMiniGauge(w, "probe pass rate", probePct)
	}

	if len(d.RPCProbes) > 0 {
		w.WriteHTML(`<div class="evm-summary__ns-row">`)
		w.WriteHTML(`<span class="evm-summary__ns-label">API</span>`)
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
		w.WriteHTML(`</div></div>`)
	}

	w.WriteHTML(`<div class="evm-summary__kpis">`)
	writeEvmSummaryKPI(w, "block", orDash(d.EVMBlock), "")
	writeEvmSummaryKPI(w, "block age", blockAge, ageTone)
	if d.EVMChainID > 0 {
		writeEvmSummaryKPI(w, "chain id", fmt.Sprintf("%d", d.EVMChainID), "")
	}
	writeEvmSummaryTxpoolKPI(w, d)
	if avg := evmAvgProbeLatency(d.RPCProbes); avg != "" {
		writeEvmSummaryKPI(w, "probe latency", avg, "")
	}
	w.WriteHTML(`</div>`)

	w.WriteHTML(fmt.Sprintf(
		`<p class="evm-summary__endpoint mono" title="%s">%s</p>`,
		html.EscapeString(httpEP), html.EscapeString(httpEP)))
	w.WriteHTML(`</div>`)
	summaryWrapEnd(w, mode)
}

func writeEvmSummaryKPI(w Writer, label, value, tone string) {
	if value == "" || value == "—" {
		return
	}
	toneClass := ""
	if tone != "" {
		toneClass = " evm-summary__kpi-val--" + tone
	}
	w.WriteHTML(fmt.Sprintf(
		`<div class="evm-summary__kpi"><span class="evm-summary__kpi-label">%s</span>`+
			`<span class="evm-summary__kpi-val%s">%s</span></div>`,
		html.EscapeString(label), toneClass, html.EscapeString(value)))
}

func writeEvmSummaryTxpoolKPI(w Writer, d model.Report) {
	pending := formatTxpoolCount(d.PendingTx, d.TxpoolGlobalSlots)
	queued := formatTxpoolCount(d.QueuedTx, d.TxpoolGlobalQueue)
	w.WriteHTML(`<div class="evm-summary__kpi evm-summary__kpi--stack">`)
	w.WriteHTML(`<span class="evm-summary__kpi-label">txpool</span>`)
	w.WriteHTML(`<span class="evm-summary__kpi-val evm-summary__kpi-val--stack">`)
	w.WriteHTML(fmt.Sprintf(`<span class="evm-summary__stack-line">%s pending</span>`,
		html.EscapeString(pending)))
	w.WriteHTML(fmt.Sprintf(`<span class="evm-summary__stack-line">%s queued</span>`,
		html.EscapeString(queued)))
	w.WriteHTML(`</span></div>`)
}

func evmBlockAgeKPI(d model.Report) (value, tone string) {
	if d.EVMBlockAge == "" {
		return "—", ""
	}
	value = d.EVMBlockAge
	switch {
	case d.EVMBlockAgeErr:
		tone = "bad"
	case d.EVMBlockAgeWarn:
		tone = "warn"
	default:
		tone = "ok"
	}
	return value, tone
}

func evmAvgProbeLatency(probes []model.RPCProbe) string {
	var sum float64
	var n int
	for _, p := range probes {
		if ms, ok := parseProbeLatencyMS(p.Latency); ok {
			sum += ms
			n++
		}
	}
	if n == 0 {
		return ""
	}
	return fmt.Sprintf("%.0fms avg", sum/float64(n))
}

func parseProbeLatencyMS(s string) (float64, bool) {
	s = strings.TrimSpace(s)
	if !strings.HasSuffix(s, "ms") {
		return 0, false
	}
	v, err := strconv.ParseFloat(strings.TrimSuffix(s, "ms"), 64)
	return v, err == nil
}

func writeEVMRPCSection(w Writer, d model.Report) {
	w.WriteHTML(evmRPCHealthCardsHTML(d))

	w.WriteHTML(`<div class="evm-wallet-section">`)
	w.WriteHTML(evmWalletCardHTML(d))
	httpEP := d.EVMHTTPEndpoint
	if httpEP == "" {
		httpEP = "http://localhost:8545"
	}
	symbol := evmDisplaySymbol(d.EVMDenom)
	networkName := strings.ToUpper(d.Network)
	if networkName == "" {
		networkName = "PMT"
	}
	wallet := fmt.Sprintf("Network name: %s\nRPC URL: %s\nChain ID: %d\nCurrency symbol: %s",
		networkName, httpEP, d.EVMChainID, symbol)
	w.Pre(wallet)
	w.WriteHTML(`</div>`)

	w.WriteHTML(`<div class="evm-probes-section">`)
	w.Subsection("Method probes")
	w.WriteHTML(evmRPCProbeTableHTML(d))
	w.WriteHTML(`</div>`)
}

func formatTxpoolCount(n, limit uint64) string {
	if limit == 0 {
		return fmt.Sprintf("%d", n)
	}
	return fmt.Sprintf("%d / %d", n, limit)
}
