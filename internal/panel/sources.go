package panel

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

const maxSourceJSONBytes = 12_000

// writeSectionSources renders the collapsible raw request/response log for a section.
func writeSectionSources(w Writer, v View, d model.Report) {
	exchanges := exchangesForView(v, d.Exchanges)
	if len(exchanges) == 0 {
		return
	}
	w.SourceLog(exchanges)
}

func exchangesForView(v View, all []model.SourceExchange) []model.SourceExchange {
	matchers := viewExchangeMatchers(v)
	if len(matchers) == 0 {
		return nil
	}
	var out []model.SourceExchange
	for _, e := range all {
		for _, m := range matchers {
			if m(e) {
				out = append(out, e)
				break
			}
		}
	}
	return out
}

func viewExchangeMatchers(v View) []func(model.SourceExchange) bool {
	switch v {
	case ViewInfra:
		return []func(model.SourceExchange) bool{
			kindMatch("file", "fs"),
			urlContains("containers/"),
		}
	case ViewNode:
		return []func(model.SourceExchange) bool{
			urlContains("/status", "/net_info", "/block", "/validators", "/num_unconfirmed", "/consensus_params"),
			urlContains("/cosmos/staking/v1beta1/validators"),
			urlContains("/cosmos/staking/v1beta1/validators/"),
			urlContains("/cosmos/bank/v1beta1/balances/"),
		}
	case ViewStaking:
		return []func(model.SourceExchange) bool{
			urlContains("/cosmos/staking/", "/cosmos/bank/", "/cosmos/auth/"),
		}
	case ViewSlashing:
		return []func(model.SourceExchange) bool{
			urlContains("/cosmos/slashing/", "/cosmos/staking/v1beta1/validators"),
		}
	case ViewFeemarket:
		return []func(model.SourceExchange) bool{
			urlContains("/feemarket/", "/block_results", "/consensus_params"),
			urlContains("app.toml"),
			kindMatch("file"),
		}
	case ViewRewards:
		return []func(model.SourceExchange) bool{
			urlContains("/cosmos/mint/", "/pmtrewards/", "/cosmos/staking/v1beta1/pool"),
		}
	case ViewDistribution:
		return []func(model.SourceExchange) bool{
			urlContains("/cosmos/distribution/", "/cosmos/bank/", "/cosmos/auth/"),
		}
	case ViewGovernance:
		return []func(model.SourceExchange) bool{
			urlContains("/cosmos/gov/", "/cosmos/upgrade/", "/ibc/", "/cosmos/evm/erc20/"),
		}
	case ViewEVM:
		return []func(model.SourceExchange) bool{
			kindMatch("jsonrpc"),
			urlContains("/cosmos/evm/vm/", "/cosmos/evm/erc20/"),
		}
	default:
		return nil
	}
}

func kindMatch(kinds ...string) func(model.SourceExchange) bool {
	set := map[string]bool{}
	for _, k := range kinds {
		set[k] = true
	}
	return func(e model.SourceExchange) bool {
		return set[e.Kind]
	}
}

func urlContains(parts ...string) func(model.SourceExchange) bool {
	return func(e model.SourceExchange) bool {
		u := strings.ToLower(e.URL)
		for _, p := range parts {
			if strings.Contains(u, strings.ToLower(p)) {
				return true
			}
		}
		return false
	}
}

func exchangeLabel(e model.SourceExchange) string {
	switch e.Kind {
	case "jsonrpc":
		if method := jsonRPCMethod(e.Request); method != "" {
			return "POST " + method
		}
		return "POST JSON-RPC"
	case "file":
		return "READ " + e.URL
	case "fs":
		return "statfs " + e.URL
	case "docker":
		return "docker " + e.Method + " " + shortenURL(e.URL)
	default:
		return e.Method + " " + shortenURL(e.URL)
	}
}

func jsonRPCMethod(request string) string {
	var req struct {
		Method string `json:"method"`
	}
	if err := json.Unmarshal([]byte(request), &req); err != nil {
		return ""
	}
	return req.Method
}

func shortenURL(u string) string {
	u = strings.TrimPrefix(u, "http://")
	u = strings.TrimPrefix(u, "https://")
	if i := strings.IndexByte(u, '/'); i >= 0 {
		return u[i:]
	}
	return u
}

func exchangeStatusLabel(ok bool) string {
	if ok {
		return "ok"
	}
	return "FAIL"
}

func prettySourceJSON(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "(empty)"
	}
	var v any
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		return truncateSourceJSON(raw)
	}
	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return truncateSourceJSON(raw)
	}
	return truncateSourceJSON(string(out))
}

func truncateSourceJSON(s string) string {
	if len(s) <= maxSourceJSONBytes {
		return s
	}
	return s[:maxSourceJSONBytes] + "\n… (truncated)"
}

func formatSourceExchange(e model.SourceExchange) string {
	var b strings.Builder
	fmt.Fprintf(&b, "── %s · %s · %s ──\n", exchangeLabel(e), exchangeStatusLabel(e.OK), e.Latency)
	if !e.OK && e.Error != "" {
		fmt.Fprintf(&b, "err » %s\n", e.Error)
	}
	if e.Request != "" && e.Request != "(none)" {
		fmt.Fprintf(&b, "req » %s\n", strings.TrimSpace(e.Request))
	} else if e.Method == "GET" || e.Method == "POST" {
		fmt.Fprintf(&b, "req » %s\n", e.Method+" "+e.URL)
	}
	b.WriteString("res » ")
	res := prettySourceJSON(e.Response)
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

func renderSourceExchangeLog(exchanges []model.SourceExchange) string {
	const (
		padLabel   = 36
		padStatus  = 6
		padLatency = 7
	)
	var b strings.Builder
	b.WriteString("  endpoint                          status  latency\n")
	b.WriteString("  ─────────────────────────────────────────────────\n")
	for _, e := range exchanges {
		status := exchangeStatusLabel(e.OK)
		mark := "·"
		if !e.OK {
			mark = "✗"
		}
		line := fmt.Sprintf("  %s  %-*s  %-*s  %-*s",
			mark, padLabel, exchangeLabel(e), padStatus, status, padLatency, e.Latency)
		if !e.OK && e.Error != "" {
			line += "  " + truncateSourceJSON(e.Error)
		}
		b.WriteString(line + "\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

func sourceExchangesHTML(exchanges []model.SourceExchange) string {
	if len(exchanges) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString(`<div class="dash-sources__log">`)
	b.WriteString(`<pre class="code-block terminal-panel dash-sources__summary-log"><code>`)
	b.WriteString(htmlEscape(renderSourceExchangeLog(exchanges)))
	b.WriteString(`</code></pre>`)
	for _, e := range exchanges {
		b.WriteString(`<pre class="code-block terminal-panel dash-sources__exchange"><code>`)
		b.WriteString(htmlEscape(formatSourceExchange(e)))
		b.WriteString(`</code></pre>`)
	}
	b.WriteString(`</div>`)
	return b.String()
}

func htmlEscape(s string) string {
	return strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
	).Replace(s)
}
