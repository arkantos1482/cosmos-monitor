package html

import (
	"fmt"
	"html"
)

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
.dash-header{
  display:flex;align-items:baseline;gap:.75rem;flex-wrap:wrap;
  margin:0 0 1.25rem;padding:.75rem 1rem;
  background:var(--surface);border:1px solid var(--border);border-radius:var(--radius);
  box-shadow:var(--shadow);
}
.dash-header h1{font-size:1.1rem;font-weight:600;letter-spacing:-.02em;color:var(--fg)}
.dash-header .meta{font-size:.8rem;color:var(--muted)}
.dash-main{display:flex;flex-direction:column;gap:1rem}
.dash-section{
  background:var(--surface);border:1px solid var(--border);border-radius:var(--radius);
  padding:1rem 1.1rem 1.15rem;box-shadow:var(--shadow);
}
.dash-heading{
  font-size:.72rem;font-weight:700;letter-spacing:.1em;text-transform:uppercase;
  color:var(--accent);margin:0 0 .85rem;padding-bottom:.5rem;border-bottom:1px solid var(--border);
}
.dash-block+.dash-block{margin-top:1rem;padding-top:1rem;border-top:1px solid var(--border)}
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
.code-block,.diagram-panel{
  background:var(--surface-2);border:1px solid var(--border);border-radius:8px;
  padding:.65rem .85rem;margin:.4rem 0 .55rem;overflow-x:auto;
}
.code-block code{font-family:ui-monospace,"Cascadia Code","Fira Code",monospace;font-size:.78rem;color:var(--fg)}
.diagram-panel.mermaid{text-align:center}
.diagram-panel.mermaid svg{max-width:100%;height:auto}
.math-display{overflow-x:auto}
.math-display .katex{font-size:1.02em}
.katex-display{margin:.45rem 0;overflow-x:auto}
code{font-family:ui-monospace,"Cascadia Code","Fira Code",monospace;font-size:.82em;color:var(--accent)}
`

// FullPage wraps an HTML fragment in the dashboard document shell.
func FullPage(moniker, fragment string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8"/>
<meta name="viewport" content="width=device-width,initial-scale=1"/>
<title>pmtop — %s</title>
<script src="https://unpkg.com/htmx.org@2.0.3/dist/htmx.min.js"></script>
<script src="https://cdn.jsdelivr.net/npm/mermaid@11/dist/mermaid.min.js"></script>
<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/katex@0.16.11/dist/katex.min.css"/>
<script defer src="https://cdn.jsdelivr.net/npm/katex@0.16.11/dist/katex.min.js"></script>
<script defer src="https://cdn.jsdelivr.net/npm/katex@0.16.11/dist/contrib/auto-render.min.js"></script>
<style>%s</style>
</head>
<body>
<header class="dash-header">
<h1>PMT operations</h1>
<span class="meta">%s · live · refreshes every 5s</span>
</header>
<main class="dash-main" id="data" hx-get="/fragment" hx-trigger="every 5s" hx-swap="innerHTML">
%s
</main>
<script>
mermaid.initialize({startOnLoad:false,theme:'dark',securityLevel:'loose'});
function renderMermaid(){
  mermaid.run({querySelector:'#data .mermaid'});
}
function renderMathDisplays(){
  if(typeof katex==='undefined')return;
  document.querySelectorAll('#data .math-display[data-tex-b64]').forEach(function(el){
    if(el.dataset.rendered)return;
    var b64=el.getAttribute('data-tex-b64');
    if(!b64)return;
    try{
      var tex=decodeURIComponent(escape(atob(b64)));
      katex.render(tex,el,{displayMode:true,throwOnError:false});
      el.dataset.rendered='1';
    }catch(e){}
  });
}
function renderMath(){
  renderMathDisplays();
  if(typeof renderMathInElement!=='function')return;
  renderMathInElement(document.getElementById('data'),{
    delimiters:[
      {left:'\\(',right:'\\)',display:false},
      {left:'\\[',right:'\\]',display:true}
    ],
    ignoredTags:['script','noscript','style','textarea','pre','code'],
    ignoredClasses:['math-display'],
    throwOnError:false
  });
}
function renderDiagrams(){renderMermaid();renderMath();}
document.addEventListener('DOMContentLoaded',renderDiagrams);
document.body.addEventListener('htmx:afterSwap',function(e){if(e.detail.target.id==='data')renderDiagrams();});
</script>
</body>
</html>`, html.EscapeString(moniker), pageCSS, html.EscapeString(moniker), fragment)
}
