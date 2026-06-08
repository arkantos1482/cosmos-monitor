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
  --bg:#0b0f14;--bg-glow:rgba(94,179,255,.07);--surface:#141b24;--surface-2:#1c2530;--border:#2a3544;
  --fg:#e8edf3;--muted:#8a9bb0;--accent:#5eb3ff;--accent-2:#5ee4a8;
  --warn:#e8c468;--bad:#f07178;--ok:#5ee4a8;
  --radius:12px;--shadow:0 1px 0 rgba(255,255,255,.04) inset,0 1px 2px rgba(0,0,0,.4),0 12px 32px rgba(0,0,0,.28);
  --font:"IBM Plex Sans",system-ui,-apple-system,sans-serif;
  --mono:"IBM Plex Mono",ui-monospace,"Cascadia Code",monospace;
}
*,*::before,*::after{box-sizing:border-box;margin:0;padding:0}
body{
  background:
    radial-gradient(ellipse 80% 50% at 50% -20%,var(--bg-glow),transparent),
    radial-gradient(ellipse 60% 40% at 100% 0%,rgba(94,228,168,.04),transparent),
    var(--bg);
  color:var(--fg);font:14px/1.6 var(--font);
  padding:1.5rem 1.75rem 3rem;max-width:1320px;margin:0 auto;
  -webkit-font-smoothing:antialiased;
}
body::before{
  content:'';position:fixed;top:0;left:0;right:0;height:2px;
  background:linear-gradient(90deg,var(--accent),var(--accent-2) 55%,var(--accent));
  z-index:10;opacity:.9;
}
.dash-shell{display:grid;grid-template-columns:210px 1fr;gap:1.5rem;align-items:start}
@media(max-width:860px){
  .dash-shell{grid-template-columns:1fr}
  .dash-nav{position:static}
}
.dash-nav{
  position:sticky;top:1.25rem;display:flex;flex-direction:column;gap:.15rem;
  padding:.7rem;background:var(--surface);border:1px solid var(--border);
  border-radius:var(--radius);box-shadow:var(--shadow);
}
.dash-nav__title{
  font-size:.65rem;font-weight:700;letter-spacing:.12em;text-transform:uppercase;
  color:var(--muted);margin:0 0 .5rem;padding:0 .45rem;
}
.dash-nav__link{
  display:block;padding:.42rem .6rem;border-radius:8px;font-size:.84rem;
  color:var(--muted);text-decoration:none;border:1px solid transparent;
  transition:color .15s,background .15s,border-color .15s;
}
.dash-nav__link:hover{color:var(--fg);background:var(--surface-2)}
.dash-nav__link--active{
  color:var(--fg);background:rgba(94,179,255,.08);border-color:rgba(94,179,255,.22);
  font-weight:600;box-shadow:inset 3px 0 0 var(--accent);
}
.dash-header{
  display:flex;align-items:center;justify-content:space-between;gap:1rem;flex-wrap:wrap;
  margin:0 0 1.25rem;padding:.85rem 1.1rem;
  background:linear-gradient(135deg,var(--surface) 0%,rgba(20,27,36,.92) 100%);
  border:1px solid var(--border);border-radius:var(--radius);box-shadow:var(--shadow);
}
.dash-header__brand{display:flex;align-items:center;gap:.65rem}
.dash-header__mark{
  width:10px;height:10px;border-radius:50%;
  background:linear-gradient(135deg,var(--accent),var(--accent-2));
  box-shadow:0 0 12px rgba(94,179,255,.55);
  animation:live-pulse 2.4s ease-in-out infinite;
}
@keyframes live-pulse{
  0%,100%{opacity:1;transform:scale(1)}
  50%{opacity:.65;transform:scale(.88)}
}
.dash-header h1{font-size:1.05rem;font-weight:600;letter-spacing:-.01em;color:var(--fg)}
.dash-header .meta{font-size:.78rem;color:var(--muted);font-variant-numeric:tabular-nums}
.meta__live{color:var(--accent-2);font-weight:500}
.dash-main{display:flex;flex-direction:column;gap:1.1rem;min-width:0}
.dash-home__lead{font-size:.86rem;color:var(--muted);margin:0 0 1rem;max-width:52ch}
.dash-cards{
  display:grid;grid-template-columns:repeat(auto-fill,minmax(250px,1fr));
  gap:.85rem;
}
.dash-card{
  display:block;padding:.9rem 1.05rem;background:var(--surface);
  border:1px solid var(--border);border-radius:var(--radius);box-shadow:var(--shadow);
  text-decoration:none;color:inherit;
  transition:border-color .2s,background .2s,transform .2s,box-shadow .2s;
}
.dash-card:hover{
  border-color:rgba(94,179,255,.35);background:var(--surface-2);
  transform:translateY(-2px);box-shadow:0 4px 20px rgba(0,0,0,.35),0 0 0 1px rgba(94,179,255,.08);
}
.dash-card__title{
  font-size:.75rem;font-weight:700;letter-spacing:.08em;text-transform:uppercase;
  color:var(--accent);margin:0 0 .5rem;
}
.dash-card__badges{margin:0 0 .45rem}
.dash-card__lines{list-style:none;margin:0;padding:0;font-size:.82rem;color:var(--fg)}
.dash-card__lines li{padding:.18rem 0;color:var(--muted);line-height:1.45}
.dash-card__lines li:first-child{color:var(--fg);font-weight:500;font-size:.84rem}
.dash-section{
  background:var(--surface);border:1px solid var(--border);border-radius:var(--radius);
  padding:1.1rem 1.2rem 1.25rem;box-shadow:var(--shadow);
}
.dash-heading{
  font-size:.7rem;font-weight:700;letter-spacing:.11em;text-transform:uppercase;
  color:var(--accent);margin:0 0 .9rem;padding-bottom:.55rem;
  border-bottom:1px solid var(--border);
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
  padding:.5rem .7rem;background:var(--surface-2);border-radius:9px;border:1px solid var(--border);
  transition:border-color .15s;
}
.stat:hover{border-color:rgba(94,179,255,.2)}
.stat dt{font-size:.7rem;font-weight:500;color:var(--muted);text-transform:lowercase}
.stat dd{font-size:.84rem;color:var(--fg);word-break:break-word;font-variant-numeric:tabular-nums}
.stat dd code{font-size:.8rem}
.badge{
  display:inline-block;padding:.14rem .5rem;border-radius:999px;
  font-size:.68rem;font-weight:600;letter-spacing:.04em;text-transform:uppercase;
  border:1px solid transparent;
}
.badge--ok{background:rgba(94,228,168,.12);color:var(--ok);border-color:rgba(94,228,168,.25)}
.badge--warn{background:rgba(232,196,104,.12);color:var(--warn);border-color:rgba(232,196,104,.25)}
.badge--bad{background:rgba(240,113,120,.12);color:var(--bad);border-color:rgba(240,113,120,.25)}
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
.data-table tbody tr:nth-child(even) td{background:rgba(255,255,255,.015)}
.data-table tbody tr:hover td{background:rgba(94,179,255,.07)}
.code-block{
  background:rgba(0,0,0,.25);border:1px solid var(--border);border-radius:9px;
  padding:.7rem .9rem;margin:.4rem 0 .55rem;overflow-x:auto;
}
.code-block code{font-family:var(--mono);font-size:.76rem;color:var(--fg);line-height:1.5}
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
.fee-meter__track{height:8px;background:rgba(0,0,0,.3);border-radius:999px;border:1px solid var(--border);overflow:hidden}
.fee-meter__fill{
  height:100%;border-radius:999px;
  background:linear-gradient(90deg,var(--accent),var(--accent-2));
  box-shadow:0 0 10px rgba(94,179,255,.35);transition:width .35s ease;
}
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
.fee-formula code{font-size:inherit;color:var(--fg);font-family:var(--mono)}
code{font-family:var(--mono);font-size:.82em;color:var(--accent)}
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
