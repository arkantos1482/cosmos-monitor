package panel

import (
	"fmt"
	"html"
	"sort"
	"strings"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

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
