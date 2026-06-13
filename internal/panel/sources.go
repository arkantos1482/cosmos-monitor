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

func truncateSourceJSON(s string) string {
	return truncateJSON(s, maxSourceJSONBytes)
}

func formatSourceExchangeHTML(e model.SourceExchange) string {
	var b strings.Builder
	b.WriteString(`<div class="dash-sources__exchange">`)
	fmt.Fprintf(&b, `<div class="dash-sources__exchange-hdr">── %s · %s · %s ──</div>`,
		htmlEscape(exchangeLabel(e)), htmlEscape(exchangeStatusLabel(e.OK)), htmlEscape(e.Latency))
	if !e.OK && e.Error != "" {
		fmt.Fprintf(&b, `<div class="dash-sources__error">err » %s</div>`, htmlEscape(truncateSourceJSON(e.Error)))
	}
	if req := sourceRequestBody(e); req != "" {
		b.WriteString(`<div class="dash-sources__payload"><span class="dash-sources__tag">req</span>`)
		if isJSON(req) {
			b.WriteString(jsonCodeBlock(req, maxSourceJSONBytes))
		} else {
			b.WriteString(plainCodeBlock(req))
		}
		b.WriteString(`</div>`)
	}
	b.WriteString(`<div class="dash-sources__payload"><span class="dash-sources__tag">res</span>`)
	b.WriteString(jsonCodeBlock(e.Response, maxSourceJSONBytes))
	b.WriteString(`</div></div>`)
	return b.String()
}

func sourceRequestBody(e model.SourceExchange) string {
	if e.Request != "" && e.Request != "(none)" {
		return strings.TrimSpace(e.Request)
	}
	if e.Method == "GET" || e.Method == "POST" {
		return e.Method + " " + e.URL
	}
	return ""
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
		b.WriteString(formatSourceExchangeHTML(e))
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
