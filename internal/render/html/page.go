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

var layoutTmpl = template.Must(template.New("layout").Parse(layoutHTML))

const pageCSS = `
:root{
  --bg:#0f1419;--surface:#1a2129;--surface-2:#242d38;--border:#2f3b4a;
  --fg:#e7ecf1;--muted:#8b9aab;--accent:#5eb3ff;--accent-2:#6ee7a8;
  --warn:#e8c468;--bad:#f07178;--ok:#6ee7a8;
  --radius:10px;--shadow:0 1px 2px rgba(0,0,0,.35),0 8px 24px rgba(0,0,0,.25);
}
*,*::before,*::after{box-sizing:border-box;margin:0;padding:0}
body{
  background:var(--bg);color:var(--fg);
  font:14px/1.55 system-ui,-apple-system,"Segoe UI",Roboto,sans-serif;
  padding:1.25rem 1.5rem 2.5rem;max-width:1280px;margin:0 auto;
}
body::before{
  content:'';position:fixed;top:0;left:0;right:0;height:3px;
  background:linear-gradient(90deg,var(--accent),var(--accent-2));z-index:10;
}
.dash-shell{display:grid;grid-template-columns:200px 1fr;gap:1.25rem;align-items:start}
@media(max-width:860px){
  .dash-shell{grid-template-columns:1fr}
  .dash-nav{position:static}
}
.dash-nav{
  position:sticky;top:1rem;display:flex;flex-direction:column;gap:.2rem;
  padding:.65rem;background:var(--surface);border:1px solid var(--border);
  border-radius:var(--radius);box-shadow:var(--shadow);
}
.dash-nav__title{
  font-size:.68rem;font-weight:700;letter-spacing:.1em;text-transform:uppercase;
  color:var(--muted);margin:0 0 .45rem;padding:0 .35rem;
}
.dash-nav__link{
  display:block;padding:.4rem .55rem;border-radius:6px;font-size:.84rem;
  color:var(--muted);text-decoration:none;border:1px solid transparent;
}
.dash-nav__link:hover{color:var(--fg);background:var(--surface-2)}
.dash-nav__link--active{
  color:var(--fg);background:var(--surface-2);border-color:var(--border);
  font-weight:600;
}
.dash-header{
  display:flex;align-items:baseline;gap:.75rem;flex-wrap:wrap;
  margin:0 0 1rem;padding:.75rem 1rem;
  background:var(--surface);border:1px solid var(--border);border-radius:var(--radius);
  box-shadow:var(--shadow);
}
.dash-header h1{font-size:1.1rem;font-weight:600;letter-spacing:-.02em;color:var(--fg)}
.dash-header .meta{font-size:.8rem;color:var(--muted)}
.dash-main{display:flex;flex-direction:column;gap:1rem;min-width:0}
.dash-home__lead{font-size:.88rem;color:var(--muted);margin:0 0 .85rem}
.dash-cards{
  display:grid;grid-template-columns:repeat(auto-fill,minmax(240px,1fr));
  gap:.75rem;
}
.dash-card{
  display:block;padding:.85rem 1rem;background:var(--surface);
  border:1px solid var(--border);border-radius:var(--radius);box-shadow:var(--shadow);
  text-decoration:none;color:inherit;transition:border-color .15s,background .15s;
}
.dash-card:hover{border-color:var(--accent);background:var(--surface-2)}
.dash-card__title{
  font-size:.82rem;font-weight:700;letter-spacing:.04em;text-transform:uppercase;
  color:var(--accent);margin:0 0 .45rem;
}
.dash-card__badges{margin:0 0 .4rem}
.dash-card__lines{list-style:none;margin:0;padding:0;font-size:.82rem;color:var(--fg)}
.dash-card__lines li{padding:.15rem 0;color:var(--muted)}
.dash-card__lines li:first-child{color:var(--fg);font-weight:500}
.dash-section{
  background:var(--surface);border:1px solid var(--border);border-radius:var(--radius);
  padding:1rem 1.1rem 1.15rem;box-shadow:var(--shadow);
}
.dash-heading{
  font-size:.72rem;font-weight:700;letter-spacing:.1em;text-transform:uppercase;
  color:var(--accent);margin:0 0 .85rem;padding-bottom:.5rem;border-bottom:1px solid var(--border);
}
.dash-block+.dash-block{margin-top:1rem;padding-top:1rem;border-top:1px solid var(--border)}
.dash-details{margin:.65rem 0 .5rem;border:1px solid var(--border);border-radius:8px;background:var(--surface-2)}
.dash-details__summary{
  cursor:pointer;padding:.55rem .75rem;font-size:.88rem;font-weight:600;color:var(--muted);
  list-style:none;user-select:none;
}
.dash-details__summary::-webkit-details-marker{display:none}
.dash-details[open] .dash-details__summary{color:var(--fg);border-bottom:1px solid var(--border)}
.dash-details__body{padding:.65rem .75rem .75rem}
.dash-details__body .dash-block:first-child{margin-top:0;padding-top:0;border-top:none}
.dash-subheading{font-size:.95rem;font-weight:600;color:var(--fg);margin:0 0 .55rem}
.hint,.note{font-size:.8rem;color:var(--muted);margin:.35rem 0 .55rem}
.hint code,.note code{font-size:.78rem}
.callout{font-size:.88rem;color:var(--fg);margin:.4rem 0 .55rem}
.callout strong{color:var(--accent-2)}
.validator-label{font-size:.85rem;font-weight:600;margin:.5rem 0 .35rem}
.stat-grid{
  display:grid;grid-template-columns:repeat(auto-fill,minmax(220px,1fr));
  gap:.45rem .75rem;margin:.25rem 0 .5rem;
}
.stat{
  display:grid;grid-template-columns:auto 1fr;gap:.35rem .65rem;align-items:baseline;
  padding:.45rem .65rem;background:var(--surface-2);border-radius:8px;border:1px solid var(--border);
}
.stat dt{font-size:.72rem;font-weight:500;color:var(--muted);text-transform:lowercase}
.stat dd{font-size:.84rem;color:var(--fg);word-break:break-word}
.stat dd code{font-size:.8rem}
.badge{
  display:inline-block;padding:.12rem .45rem;border-radius:999px;
  font-size:.72rem;font-weight:600;letter-spacing:.02em;text-transform:uppercase;
}
.badge--ok{background:rgba(110,231,168,.15);color:var(--ok)}
.badge--warn{background:rgba(232,196,104,.15);color:var(--warn)}
.badge--bad{background:rgba(240,113,120,.15);color:var(--bad)}
.dash-list{list-style:none;margin:.35rem 0 .5rem;padding:0}
.dash-list li{
  padding:.35rem .55rem;margin:.2rem 0;background:var(--surface-2);
  border-radius:6px;border:1px solid var(--border);font-size:.84rem;
}
.table-scroll{overflow-x:auto;margin:.45rem 0 .65rem;border-radius:8px;border:1px solid var(--border)}
.data-table{border-collapse:collapse;width:100%;font-size:.82rem}
.data-table thead th{
  text-align:left;color:var(--muted);font-weight:500;padding:.45rem .65rem;
  background:var(--surface-2);border-bottom:1px solid var(--border);white-space:nowrap;
}
.data-table tbody td{padding:.4rem .65rem;border-bottom:1px solid var(--border);white-space:nowrap}
.data-table tbody tr:last-child td{border-bottom:none}
.data-table tbody tr:hover td{background:rgba(94,179,255,.06)}
.code-block{
  background:var(--surface-2);border:1px solid var(--border);border-radius:8px;
  padding:.65rem .85rem;margin:.4rem 0 .55rem;overflow-x:auto;
}
.code-block code{font-family:ui-monospace,"Cascadia Code","Fira Code",monospace;font-size:.78rem;color:var(--fg)}
.fee-traffic{margin:.5rem 0 .75rem}
.fee-badge{
  display:inline-block;padding:.45rem 1rem;border-radius:8px;
  font-size:.85rem;font-weight:700;letter-spacing:.06em;margin-bottom:.65rem;
}
.fee-badge--rising{background:rgba(240,113,120,.18);color:var(--bad);border:1px solid rgba(240,113,120,.35)}
.fee-badge--falling{background:rgba(110,231,168,.15);color:var(--ok);border:1px solid rgba(110,231,168,.35)}
.fee-badge--stable{background:rgba(94,179,255,.12);color:var(--accent);border:1px solid rgba(94,179,255,.3)}
.fee-meter{margin:.35rem 0 .55rem}
.fee-meter-note{margin:.35rem 0 .55rem;font-size:.8rem;color:var(--muted);line-height:1.4}
.fee-meter__label{display:flex;justify-content:space-between;font-size:.75rem;color:var(--muted);margin-bottom:.3rem}
.fee-meter__track{height:10px;background:var(--surface-2);border-radius:999px;border:1px solid var(--border);overflow:hidden}
.fee-meter__fill{height:100%;border-radius:999px;background:linear-gradient(90deg,var(--accent),var(--accent-2));transition:width .3s ease}
.fee-hero-line{margin:.35rem 0 .5rem;font-size:.88rem;line-height:1.45}
.fee-key-metrics{margin:.35rem 0 .65rem}
.fee-cards{
  display:grid;grid-template-columns:repeat(auto-fit,minmax(260px,1fr));
  gap:.65rem;margin:.55rem 0 .65rem;
}
.fee-card{
  padding:.65rem .75rem;background:var(--surface-2);border:1px solid var(--border);border-radius:8px;
}
.fee-card__title{font-size:.8rem;font-weight:600;color:var(--muted);text-transform:uppercase;letter-spacing:.05em;margin:0 0 .45rem}
.fee-formula{
  background:var(--surface-2);border:1px solid var(--border);border-radius:8px;
  padding:.45rem .65rem;margin:.35rem 0;font-size:.78rem;line-height:1.45;
  overflow-x:auto;white-space:pre;
}
.fee-formula code{font-size:inherit;color:var(--text);font-family:ui-monospace,"Cascadia Code","Fira Code",monospace}
code{font-family:ui-monospace,"Cascadia Code","Fira Code",monospace;font-size:.82em;color:var(--accent)}
`

type pageData struct {
	Moniker   string
	PageTitle string
	Nav       template.HTML
	Content   template.HTML
	DataURL   string
	CSS       template.CSS
}

func navHTML(active panel.View) string {
	var b strings.Builder
	fmt.Fprint(&b, `<nav id="dash-nav" class="dash-nav" aria-label="Sections">`)
	fmt.Fprint(&b, `<p class="dash-nav__title">Sections</p>`)
	for _, item := range panel.Nav {
		cls := "dash-nav__link"
		if item.View == active {
			cls += " dash-nav__link--active"
		}
		fmt.Fprintf(&b, `<a class="%s" href="%s">%s</a>`,
			cls, html.EscapeString(item.Path), html.EscapeString(item.Label))
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
func FullPage(moniker string, active panel.View, fragment string) string {
	var buf bytes.Buffer
	_ = layoutTmpl.Execute(&buf, pageData{
		Moniker:   html.EscapeString(moniker),
		PageTitle: html.EscapeString(pageTitle(active)),
		Nav:       template.HTML(navHTML(active)),
		Content:   template.HTML(fragment),
		DataURL:   dataURL(active),
		CSS:       pageCSS,
	})
	return buf.String()
}
