package html

import (
	"bytes"
	_ "embed"
	"fmt"
	"html"
	"html/template"
	"strings"

	"github.com/arkantos1482/cosmos-monitor/internal/panel"
)

//go:embed templates/layout.html
var layoutHTML string

//go:embed static/theme.css
var themeCSS string

var layoutTmpl = template.Must(template.New("layout").Parse(layoutHTML))

type pageData struct {
	Moniker   string
	PageTitle string
	Status    template.HTML
	Nav       template.HTML
	Content   template.HTML
	DataURL   string
	CSS       template.CSS
}

var navIcons = map[panel.View]string{
	panel.ViewHome: `<svg class="dash-nav__icon" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true"><path d="M2 6.5L8 2l6 4.5V13a1 1 0 01-1 1H3a1 1 0 01-1-1V6.5z"/><path d="M6 14V9h4v5"/></svg>`,
	panel.ViewInfra: `<svg class="dash-nav__icon" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true"><rect x="2" y="3" width="12" height="10" rx="1"/><path d="M5 7h6M5 10h4"/></svg>`,
	panel.ViewNode: `<svg class="dash-nav__icon" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true"><circle cx="8" cy="8" r="5"/><path d="M8 5v3l2 1"/></svg>`,
	panel.ViewStaking: `<svg class="dash-nav__icon" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true"><path d="M8 2l2 3.5h4l-3.2 2.5 1.2 4L8 10.5 3.8 12l1.2-4L2 5.5h4L8 2z"/></svg>`,
	panel.ViewRewards: `<svg class="dash-nav__icon" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true"><path d="M8 2l1.8 3.6L14 6.3l-3 2.9.7 4.1L8 11.2 4.3 13.3 5 9.2 2 6.3l4.2-.7L8 2z"/></svg>`,
	panel.ViewFeemarket: `<svg class="dash-nav__icon" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true"><path d="M3 12h10"/><path d="M8 3l4 6H4l4-6z"/><circle cx="8" cy="11" r="1"/></svg>`,
	panel.ViewGovernance: `<svg class="dash-nav__icon" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true"><path d="M3 4h10M3 8h10M3 12h6"/></svg>`,
	panel.ViewEVM: `<svg class="dash-nav__icon" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true"><path d="M2 8l3-5 3 5-3 5-3-5zM8 8l3-5 3 5-3 5-3-5z"/></svg>`,
}

func navSlug(v panel.View) string {
	switch v {
	case panel.ViewInfra:
		return "infra"
	case panel.ViewNode:
		return "node"
	case panel.ViewStaking:
		return "staking"
	case panel.ViewRewards:
		return "rewards"
	case panel.ViewFeemarket:
		return "feemarket"
	case panel.ViewGovernance:
		return "governance"
	case panel.ViewEVM:
		return "evm"
	default:
		return "home"
	}
}

func writeNavLink(b *strings.Builder, item panel.NavItem, active panel.View) {
	cls := "dash-nav__link dash-nav__link--" + navSlug(item.View)
	if item.View == active {
		cls += " dash-nav__link--active"
	}
	icon := navIcons[item.View]
	fmt.Fprintf(b, `<a class="%s" href="%s">%s%s</a>`,
		cls, html.EscapeString(item.Path), icon, html.EscapeString(item.Label))
}

func navHTML(active panel.View) string {
	var b strings.Builder
	fmt.Fprint(&b, `<nav id="dash-nav" class="dash-nav" aria-label="Sections">`)
	for _, item := range panel.Nav {
		if item.View == panel.ViewHome {
			fmt.Fprint(&b, `<p class="dash-nav__title">Overview</p>`)
			writeNavLink(&b, item, active)
			break
		}
	}
	var lastScope panel.NavScope
	for _, item := range panel.Nav {
		if item.Scope == "" {
			continue
		}
		if item.Scope != lastScope {
			fmt.Fprintf(&b, `<p class="dash-nav__group">%s</p>`, html.EscapeString(panel.NavScopeLabel(item.Scope)))
			lastScope = item.Scope
		}
		writeNavLink(&b, item, active)
	}
	fmt.Fprint(&b, `</nav>`)
	return b.String()
}

func pageTitle(active panel.View) string {
	for _, item := range panel.Nav {
		if item.View == active {
			return item.Label
		}
	}
	return "Overview"
}

func dataURL(active panel.View) string {
	if active == panel.ViewHome {
		return "/"
	}
	return "/s/" + string(active)
}

// FullPage wraps an HTML fragment in the dashboard document shell.
func FullPage(moniker string, active panel.View, statusStrip, fragment string) string {
	var buf bytes.Buffer
	_ = layoutTmpl.Execute(&buf, pageData{
		Moniker:   html.EscapeString(moniker),
		PageTitle: html.EscapeString(pageTitle(active)),
		Status:    template.HTML(statusStrip),
		Nav:       template.HTML(navHTML(active)),
		Content:   template.HTML(fragment),
		DataURL:   dataURL(active),
		CSS:       template.CSS(themeCSS),
	})
	return buf.String()
}
