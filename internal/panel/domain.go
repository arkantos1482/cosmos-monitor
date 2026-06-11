package panel

import (
	"fmt"
	"html"
	"strings"
)

func orEcoDash(s string) string {
	if s == "" || s == "—" {
		return "—"
	}
	return s
}

func ecoDomainValueHTML(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "true":
		return `<span class="badge badge--ok">true</span>`
	case "false":
		return `<span class="badge badge--bad">false</span>`
	default:
		return softWrapHTML(v)
	}
}

func ecoDomainRow(b *strings.Builder, rowClass, param, value, effect string) {
	ecoDomainRowHTML(b, rowClass, param, ecoDomainValueHTML(value), effect)
}

func ecoDomainRowHTML(b *strings.Builder, rowClass, param, valueHTML, effect string) {
	cls := `eco-domain__row`
	if mod := strings.Trim(rowClass, ` "'`); mod != "" {
		cls += " " + mod
	}
	fmt.Fprintf(b, `<div class="%s"><div class="eco-domain__param">%s</div><div class="eco-domain__value">%s</div><div class="eco-domain__effect">%s</div></div>`,
		cls, html.EscapeString(param), valueHTML, html.EscapeString(effect))
}

func ecoBalanceAddrHTML(balance, addr string) string {
	bal := strings.TrimSpace(balance)
	if bal == "" || bal == "—" {
		bal = ""
	}
	addr = strings.TrimSpace(addr)
	if bal == "" && addr == "" {
		return "—"
	}
	var b strings.Builder
	b.WriteString(`<div class="eco-acct">`)
	if bal != "" {
		fmt.Fprintf(&b, `<div class="eco-acct__balance">%s</div>`, html.EscapeString(bal))
	}
	if addr != "" {
		fmt.Fprintf(&b, `<code class="eco-acct__addr">%s</code>`, html.EscapeString(addr))
	}
	b.WriteString(`</div>`)
	return b.String()
}

func ecoDomainDivider(b *strings.Builder) {
	b.WriteString(`<div class="eco-domain__divider">Governance params</div>`)
}

func ecoDomainCardOpen(b *strings.Builder, modClass, title, subtitle string) {
	fmt.Fprintf(b, `<div class="eco-domain %s">`, modClass)
	fmt.Fprintf(b, `<h3 class="eco-domain__title">%s <span class="eco-domain__subtitle">%s</span></h3>`,
		html.EscapeString(title), html.EscapeString(subtitle))
	b.WriteString(`<div class="eco-domain__rows">`)
}

func ecoDomainCardClose(b *strings.Builder) {
	b.WriteString(`</div></div>`)
}

func ecoDomainsWrap(cards ...string) string {
	var b strings.Builder
	b.WriteString(`<div class="eco-domains">`)
	for _, c := range cards {
		b.WriteString(c)
	}
	b.WriteString(`</div>`)
	return b.String()
}
