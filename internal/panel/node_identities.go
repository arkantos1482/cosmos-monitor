package panel

import (
	"fmt"
	"html"
	"strings"

	"github.com/arkantos1482/cosmos-monitor/internal/fetch"
	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

type identityRow struct {
	role   string
	bech32 string
	hex    string
}

func localConsensusBech32(lv model.LocalValidator) string {
	consBech := lv.ConsensusBech32
	if consBech == "" && lv.ConsensusAddr != "" {
		consBech = fetch.ConsHexToBech32(lv.ConsensusAddr)
	}
	return consBech
}

func writeStakingDelegators(w Writer, lv model.LocalValidator) {
	if len(lv.Delegations) == 0 {
		w.Hint("No delegations returned for this validator.")
	} else {
		w.WriteHTML(delegationsTableHTML(lv.Delegations))
	}
	w.Hint("`address`, `delegated`, `shares` → REST GET /cosmos/staking/v1beta1/validators/{valoper}/delegations; " +
		"`liquid` → REST GET /cosmos/bank/v1beta1/balances/{address}.")
}

func identityDualAddrHTML(bech32, evm string) string {
	bech32 = strings.TrimSpace(bech32)
	evm = strings.TrimSpace(evm)
	if bech32 == "" && evm == "" {
		return `<span class="id-empty">—</span>`
	}
	var b strings.Builder
	b.WriteString(`<div class="id-dual">`)
	if bech32 != "" {
		b.WriteString(`<div class="id-dual__bech32">`)
		b.WriteString(identityBech32Cell(bech32, ""))
		b.WriteString(`</div>`)
	}
	if evm != "" {
		fmt.Fprintf(&b, `<code class="id-dual__evm">%s</code>`, html.EscapeString(evm))
	}
	b.WriteString(`</div>`)
	return b.String()
}

func delegationsTableHTML(delegations []model.DelegationRow) string {
	var b strings.Builder
	b.WriteString(`<div class="table-scroll table-scroll--fit"><table class="data-table data-table--delegations"><thead><tr>`)
	for _, h := range []string{"address", "delegated", "liquid", "shares"} {
		thCls := tableColumnClass(h)
		fmt.Fprintf(&b, "<th%s>%s</th>", thCls, html.EscapeString(h))
	}
	b.WriteString(`</tr></thead><tbody>`)
	for _, d := range delegations {
		trAttr := ""
		if d.IsLocal {
			trAttr = fmt.Sprintf(` class="%s" title="this node"`, validatorLocalRowClass)
		}
		fmt.Fprintf(&b, "<tr%s>", trAttr)
		fmt.Fprintf(&b, `<td>%s</td>`, identityDualAddrHTML(d.Delegator, d.EVMAddr))
		for _, col := range []struct {
			header string
			val    string
		}{
			{"delegated", d.Balance},
			{"liquid", orLiquidDash(d.LiquidBalance)},
			{"shares", orSharesDash(d.Shares)},
		} {
			fmt.Fprintf(&b, "<td%s>%s</td>", tableColumnClass(col.header), formatValue(col.val))
		}
		b.WriteString(`</tr>`)
	}
	b.WriteString(`</tbody></table></div>`)
	return b.String()
}

func orLiquidDash(s string) string {
	if strings.TrimSpace(s) == "" {
		return "—"
	}
	return s
}

func orSharesDash(s string) string {
	s = strings.TrimSpace(s)
	if s == "" || s == "0" {
		return "—"
	}
	return s
}

func validatorIdentityBoardHTML(d model.Report, lv model.LocalValidator) string {
	return identityBoardHTML(d, lv, []identityRow{
		{role: "consensus", bech32: localConsensusBech32(lv), hex: formatConsensusHex(lv.ConsensusAddr)},
		{role: "p2p", bech32: "", hex: strings.ToLower(d.NodeID)},
	}, "")
}

func identityBoardHTML(d model.Report, lv model.LocalValidator, rows []identityRow, sharedStem string) string {
	if len(rows) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString(`<div class="id-board">`)
	if moniker := strings.TrimSpace(lv.Moniker); moniker != "" {
		fmt.Fprintf(&b, `<div class="id-board__moniker">%s</div>`, html.EscapeString(moniker))
	} else if m := strings.TrimSpace(d.Moniker); m != "" {
		fmt.Fprintf(&b, `<div class="id-board__moniker">%s</div>`, html.EscapeString(m))
	}
	b.WriteString(`<div class="table-scroll"><table class="data-table id-board__table"><thead><tr>`)
	b.WriteString(`<th class="id-board__role"></th><th>bech32</th><th>hex</th></tr></thead><tbody>`)
	for _, row := range rows {
		stem := ""
		if row.role == "account" || row.role == "operator" {
			stem = sharedStem
		}
		fmt.Fprintf(&b, `<tr class="id-board__row id-board__row--%s">`, html.EscapeString(row.role))
		fmt.Fprintf(&b, `<td class="id-board__role">%s</td>`, html.EscapeString(row.role))
		fmt.Fprintf(&b, `<td class="id-board__bech32">%s</td>`, identityBech32Cell(row.bech32, stem))
		fmt.Fprintf(&b, `<td class="id-board__hex">%s</td>`, identityHexCell(row.role, row.hex))
		b.WriteString(`</tr>`)
	}
	b.WriteString(`</tbody></table></div></div>`)
	return b.String()
}

func identityBech32Cell(addr, sharedStem string) string {
	if addr == "" {
		return `<span class="id-empty">—</span>`
	}
	hrp, data := splitBech32HRP(addr)
	if hrp == "" {
		return `<code>` + html.EscapeString(addr) + `</code>`
	}
	var out strings.Builder
	fmt.Fprintf(&out, `<code><span class="id-hrp">%s</span><span class="id-sep">1</span>`, html.EscapeString(hrp))
	out.WriteString(formatBech32Data(data, sharedStem))
	out.WriteString(`</code>`)
	return out.String()
}

func identityHexCell(role, hexVal string) string {
	if hexVal == "" {
		return `<span class="id-empty">—</span>`
	}
	cls := "id-hex"
	switch role {
	case "account":
		cls += " id-hex--evm"
	case "consensus":
		cls += " id-hex--cons"
	case "p2p":
		cls += " id-hex--p2p"
	}
	return fmt.Sprintf(`<code class="%s">%s</code>`, cls, html.EscapeString(hexVal))
}

func formatBech32Data(data, sharedStem string) string {
	if sharedStem != "" && strings.HasPrefix(data, sharedStem) {
		shared := html.EscapeString(sharedStem)
		rest := html.EscapeString(data[len(sharedStem):])
		return `<span class="id-shared">` + shared + `</span>` + rest
	}
	return html.EscapeString(data)
}

func formatConsensusHex(addr string) string {
	addr = strings.TrimPrefix(strings.TrimSpace(addr), "0x")
	addr = strings.TrimPrefix(addr, "0X")
	if addr == "" {
		return ""
	}
	return strings.ToUpper(addr)
}

func splitBech32HRP(addr string) (hrp, data string) {
	for _, p := range []string{
		fetch.Bech32PrefixValOper,
		fetch.Bech32PrefixCons,
		fetch.Bech32PrefixAcc,
	} {
		if strings.HasPrefix(addr, p+"1") {
			return p, addr[len(p)+1:]
		}
	}
	return "", addr
}

func bech32DataPart(addr string) string {
	_, data := splitBech32HRP(addr)
	return data
}

func longestCommonPrefix(a, b string) string {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		if a[i] != b[i] {
			return a[:i]
		}
	}
	return a[:n]
}
