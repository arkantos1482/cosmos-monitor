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
	if dw, ok := w.(*docWriter); ok {
		if dw.inSection {
			dw.SourceLog(exchanges)
			return
		}
		dw.emitSources(sourcesSlugForView(v), exchanges)
		return
	}
	w.SourceLog(exchanges)
}

func sourcesSlugForView(v View) string {
	switch v {
	case ViewHome:
		return "overview"
	default:
		return string(v)
	}
}

func exchangesForView(v View, all []model.SourceExchange) []model.SourceExchange {
	if v == ViewHome {
		return all
	}
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
			urlContains("/status", "/net_info", "/block", "/validators", "/num_unconfirmed"),
			urlContains("/cosmos/staking/v1beta1/validators"),
			urlContains("/cosmos/slashing/v1beta1/signing_infos"),
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
			urlContains("/cosmos/evm/vm/v1/params"),
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
		if e.Method == "WS" {
			if method := jsonRPCMethod(e.Request); method != "" {
				return "WS " + method
			}
			return "WS JSON-RPC"
		}
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

func exchangeEndpointParts(e model.SourceExchange) (verb, path string) {
	switch e.Kind {
	case "jsonrpc":
		if e.Method == "WS" {
			if method := jsonRPCMethod(e.Request); method != "" {
				return "WS", method
			}
			return "WS", "JSON-RPC"
		}
		if method := jsonRPCMethod(e.Request); method != "" {
			return "POST", method
		}
		return "POST", "JSON-RPC"
	case "file":
		return "READ", e.URL
	case "fs":
		return "statfs", e.URL
	case "docker":
		return "docker " + e.Method, shortenURL(e.URL)
	default:
		return e.Method, shortenURL(e.URL)
	}
}

func renderSourceExchangeTable(exchanges []model.SourceExchange) string {
	var b strings.Builder
	b.WriteString(`<table class="dash-sources__table"><thead><tr>`)
	b.WriteString(`<th class="dash-sources__mark-hdr" aria-hidden="true"></th>`)
	b.WriteString(`<th>endpoint</th><th>status</th><th>latency</th>`)
	b.WriteString(`</tr></thead><tbody>`)
	for _, e := range exchanges {
		verb, path := exchangeEndpointParts(e)
		status := exchangeStatusLabel(e.OK)
		mark := "·"
		rowClass := ""
		if !e.OK {
			mark = "✗"
			rowClass = ` class="dash-sources__row--fail"`
		}
		fmt.Fprintf(&b, `<tr%s><td class="dash-sources__mark">%s</td><td class="dash-sources__endpoint"><div class="dash-sources__endpoint-inner">`, rowClass, mark)
		fmt.Fprintf(&b, `<span class="dash-sources__verb">%s</span>`, htmlEscape(verb))
		if path != "" {
			fmt.Fprintf(&b, `<span class="dash-sources__path">%s</span>`, htmlEscape(path))
		}
		b.WriteString(`</div></td>`)
		fmt.Fprintf(&b, `<td class="dash-sources__status">%s</td>`, htmlEscape(status))
		fmt.Fprintf(&b, `<td class="dash-sources__latency">%s</td></tr>`, htmlEscape(e.Latency))
	}
	b.WriteString(`</tbody></table>`)
	return b.String()
}

func sourceExchangesHTML(exchanges []model.SourceExchange) string {
	if len(exchanges) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString(`<div class="dash-sources__log">`)
	b.WriteString(renderSourceExchangeTable(exchanges))
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
