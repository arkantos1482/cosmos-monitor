package main

import (
	"bytes"
	"fmt"
	"html"
	"log"
	"net/http"

	"github.com/arkantos1482/cosmos-monitor/fetch"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

var mdRenderer = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
)

func renderFragment(d WebData) string {
	src := buildMarkdown(d)
	src, mathBlocks := stripDisplayMathForGoldmark(src)
	var buf bytes.Buffer
	if err := mdRenderer.Convert([]byte(src), &buf); err != nil {
		return "<pre>" + html.EscapeString(src) + "</pre>"
	}
	return injectDisplayMathHTML(buf.String(), mathBlocks)
}

func fullPage(moniker, fragment string) string {
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
<div id="data" hx-get="/fragment" hx-trigger="every 5s" hx-swap="innerHTML">
%s
</div>
<script>
mermaid.initialize({startOnLoad:false,theme:'dark',securityLevel:'loose'});
function promoteMermaidFences(){
  document.querySelectorAll('#data pre > code.language-mermaid').forEach(function(code){
    if(code.dataset.promoted)return;
    var div=document.createElement('div');
    div.className='mermaid';
    div.textContent=code.textContent;
    code.parentElement.replaceWith(div);
    code.dataset.promoted='1';
  });
}
function renderMermaid(){
  promoteMermaidFences();
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
</html>`, html.EscapeString(moniker), pageCSS, fragment)
}

const pageCSS = `
:root{--bg:#0d1117;--surface:#161b22;--border:#30363d;--fg:#c9d1d9;--dim:#6e7681;--bright:#f0f6fc;--red:#ff7b72;--green:#3fb950;--yellow:#e3b341;--cyan:#79c0ff}
*,*::before,*::after{box-sizing:border-box;margin:0;padding:0}
body{background:var(--bg);color:var(--fg);font-family:'Cascadia Code','Fira Code','JetBrains Mono',ui-monospace,monospace;font-size:13px;line-height:1.6;padding:1.5rem 2rem;max-width:1400px;margin:0 auto}
body::before{content:'';position:fixed;top:0;left:0;right:0;height:2px;background:linear-gradient(90deg,var(--cyan),var(--green));z-index:10}
h1{color:var(--cyan);font-size:11px;font-weight:700;text-transform:uppercase;letter-spacing:.08em;margin:1.5rem 0 .4rem;padding-bottom:.3rem;border-bottom:1px solid var(--border)}
h2{color:var(--yellow);font-size:11px;font-weight:600;margin:.8rem 0 .25rem 1rem}
ul{list-style:none;padding:0 1.5rem;margin:0 0 .25rem}
li{padding:1px 0}
li>p{display:inline;margin:0}
li strong{color:var(--dim);font-weight:400}
table{border-collapse:collapse;width:100%;margin:.5rem 0 1rem;font-size:12px;overflow-x:auto;display:block}
thead th{text-align:left;color:var(--dim);font-weight:500;padding:4px 10px;border-bottom:1px solid var(--border);font-size:11px;white-space:nowrap}
tbody td{padding:4px 10px;border-bottom:1px solid #1c2128;white-space:nowrap}
tbody tr:last-child td{border-bottom:none}
tbody tr:hover td{background:#1c2128}
pre{background:var(--surface);border:1px solid var(--border);border-radius:4px;padding:.6rem 1rem;margin:.4rem 0 .6rem;overflow-x:auto}
.mermaid{background:var(--surface);border:1px solid var(--border);border-radius:4px;padding:.6rem;margin:.4rem 0 .6rem;overflow-x:auto;text-align:center}
.mermaid svg{max-width:100%;height:auto}
.math-display{background:var(--surface);border:1px solid var(--border);border-radius:4px;padding:.65rem 1rem;margin:.4rem 0 .6rem;overflow-x:auto}
.math-display .katex{font-size:1.05em}
.katex-display{margin:.5rem 0;overflow-x:auto}
code{font-family:inherit;font-size:12px;color:var(--cyan);background:transparent}
pre code{color:var(--fg)}
p{margin:.3rem 0;color:var(--dim);font-size:12px}
em{font-style:normal;color:var(--dim)}
`

func startWeb(addr string, evmEndpoint string, doFetch func() (fetch.ChainSnapshot, fetch.EVMSnapshot, fetch.SystemSnapshot, fetch.DockerSnapshot)) {
	render := func() (WebData, string) {
		chain, ev, sys, docker := doFetch()
		d := buildWebData(chain, ev, sys, docker, evmEndpoint)
		return d, renderFragment(d)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		d, fragment := render()
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, fullPage(d.Moniker, fragment))
	})

	http.HandleFunc("/fragment", func(w http.ResponseWriter, r *http.Request) {
		_, fragment := render()
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, fragment)
	})

	go func() {
		log.Printf("web UI → http://localhost%s", addr)
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Printf("web server: %v", err)
		}
	}()
}
