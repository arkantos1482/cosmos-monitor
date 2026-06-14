package panel

import (
	"fmt"
	"html"
	"strconv"
	"strings"
)

func orDash(s string) string {
	if strings.TrimSpace(s) == "" {
		return "—"
	}
	return s
}

func boolStr(v bool) string {
	if v {
		return "true"
	}
	return "false"
}

func formatParamInt(n int64) string {
	return strconv.FormatInt(n, 10)
}

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
	cls := "eco-domain__row"
	if mod := ecoRowModifierClass(rowClass); mod != "" {
		cls += " " + mod
	}
	fmt.Fprintf(b, `<div class="%s"><div class="eco-domain__param">%s</div><div class="eco-domain__value">%s</div><div class="eco-domain__effect">%s</div></div>`,
		cls, html.EscapeString(param), valueHTML, html.EscapeString(effect))
}

// ecoRowModifierClass accepts either "eco-domain__row--warn" or legacy ` class="eco-domain__row--warn"`.
func ecoRowModifierClass(rowClass string) string {
	s := strings.TrimSpace(rowClass)
	if strings.HasPrefix(s, `class="`) {
		s = strings.TrimPrefix(s, `class="`)
		if i := strings.IndexByte(s, '"'); i >= 0 {
			s = s[:i]
		}
	}
	return strings.Trim(s, `"' `)
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
	ecoDomainCardTitle(b, title, subtitle, "", "")
	b.WriteString(`<div class="eco-domain__rows">`)
}

func ecoDomainCardTitle(b *strings.Builder, title, subtitle, badgeClass, statusLabel string) {
	b.WriteString(`<h3 class="eco-domain__title">`)
	b.WriteString(html.EscapeString(title))
	fmt.Fprintf(b, ` <span class="eco-domain__subtitle">%s</span>`, html.EscapeString(subtitle))
	if statusLabel != "" {
		fmt.Fprintf(b, ` <span class="eco-domain__status badge %s">%s</span>`,
			html.EscapeString(badgeClass), html.EscapeString(statusLabel))
	}
	b.WriteString(`</h3>`)
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
