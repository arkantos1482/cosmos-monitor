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

func writeIdentityBoard(w Writer, d model.Report, lv model.LocalValidator) {
	w.WriteHTML(identityBoardHTML(d, lv))
	w.Hint("`account`, `operator` → x/staking + GET /cosmos/evm/vm/v1/validator_account/{cons_address}; `consensus` → x/slashing / staking pubkey; `p2p` → CometBFT GET /status node_info.id.")
}

func identityBoardHTML(d model.Report, lv model.LocalValidator) string {
	consBech := lv.ConsensusBech32
	if consBech == "" && lv.ConsensusAddr != "" {
		consBech = fetch.ConsHexToBech32(lv.ConsensusAddr)
	}
	consHex := formatConsensusHex(lv.ConsensusAddr)

	accountBech := lv.AccountAddr
	accountHex := lv.EVMAddr
	if accountHex == "" && accountBech != "" {
		accountHex = fetch.AccBech32ToEVM(accountBech)
	}

	rows := []identityRow{
		{role: "account", bech32: accountBech, hex: accountHex},
		{role: "operator", bech32: lv.OperatorAddr, hex: ""},
		{role: "consensus", bech32: consBech, hex: consHex},
		{role: "p2p", bech32: "", hex: strings.ToLower(d.NodeID)},
	}

	accData := bech32DataPart(rows[0].bech32)
	opData := bech32DataPart(rows[1].bech32)
	sharedStem := longestCommonPrefix(accData, opData)
	if len(sharedStem) < 8 {
		sharedStem = ""
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
