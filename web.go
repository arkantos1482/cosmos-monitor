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
	gmhtml "github.com/yuin/goldmark/renderer/html"
)

var mdRenderer = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
	goldmark.WithRendererOptions(gmhtml.WithUnsafe()),
)

func renderFragment(d WebData) string {
	src := buildMarkdown(d, true)
	var buf bytes.Buffer
	if err := mdRenderer.Convert([]byte(src), &buf); err != nil {
		return "<pre>" + html.EscapeString(src) + "</pre>"
	}
	return buf.String()
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
function renderMermaid(){mermaid.run({querySelector:'#data .mermaid'});}
function renderFeeMathTex(){
  if(typeof katex==='undefined')return;
  document.querySelectorAll('#data .fee-math-tex').forEach(function(el){
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
  renderFeeMathTex();
  if(typeof renderMathInElement!=='function')return;
  renderMathInElement(document.getElementById('data'),{
    delimiters:[
      {left:'$$',right:'$$',display:true},
      {left:'\\(',right:'\\)',display:false},
      {left:'\\[',right:'\\]',display:true}
    ],
    ignoredTags:['script','noscript','style','textarea','pre','code'],
    ignoredClasses:['fee-math'],
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
.fee-math{background:var(--surface);border:1px solid var(--border);border-radius:4px;padding:.8rem 1rem;margin:.4rem 0 .6rem;overflow-x:auto}
.fee-math .katex{font-size:1.05em}
.fee-math-tex{margin:.35rem 0}
code{font-family:inherit;font-size:12px;color:var(--cyan);background:transparent}
pre code{color:var(--fg)}
p{margin:.3rem 0;color:var(--dim);font-size:12px}
em{font-style:normal;color:var(--dim)}
.evm-rpc-strip{display:flex;flex-wrap:wrap;gap:.35rem;margin:.5rem 0 1rem;padding:.5rem .65rem;background:var(--surface);border:1px solid var(--border);border-radius:6px}
.evm-pill{display:inline-block;padding:2px 8px;border-radius:4px;font-size:11px;font-weight:600;color:var(--bright);background:#21262d;border:1px solid var(--border)}
.evm-pill-ok{border-color:var(--green);color:var(--green)}
.evm-pill-warn{border-color:var(--yellow);color:var(--yellow)}
.evm-pill-err{border-color:var(--red);color:var(--red)}
.evm-wallet-snippet{margin:.4rem 0 .8rem}
.evm-wallet-snippet pre{background:var(--surface);border:1px solid var(--border);border-radius:4px;padding:.6rem 1rem;margin:0;color:var(--fg);font-size:12px}
.evm-probe-log{background:var(--surface);border:1px solid var(--border);border-radius:4px;padding:.65rem 1rem;margin:.45rem 0 .75rem;font-size:12px;line-height:1.55;color:var(--fg);overflow-x:auto;white-space:pre;tab-size:4}
.evm-probe-fail-head{margin:.5rem 0 .15rem;padding:.35rem .75rem;background:#2d1f1f;border:1px solid #6e3030;border-radius:4px 4px 0 0;color:var(--red);font-size:12px}
.evm-probe-fail-err{margin:0 0 .15rem;padding:.25rem .75rem;background:#1c1414;border-left:1px solid #6e3030;border-right:1px solid #6e3030;color:var(--red);font-size:11px}
.evm-probe-cmd,.evm-probe-json{margin:0 0 .5rem;padding:.5rem .75rem;background:#0d1117;border:1px solid var(--border);border-radius:0;font-size:11px;line-height:1.5;overflow-x:auto;white-space:pre-wrap;word-break:break-word}
.evm-probe-json{border-radius:0 0 4px 4px;margin-bottom:.75rem;color:var(--dim)}
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
