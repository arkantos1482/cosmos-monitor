package main

import (
	"bytes"
	"fmt"
	"html"
	"log"
	"net/http"

	"github.com/arkantos1482/cosmos-monitor/fetch"
)

const page = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<title>pmtop</title>
<script src="https://unpkg.com/htmx.org@2.0.3/dist/htmx.min.js"></script>
<style>
*{box-sizing:border-box;margin:0;padding:0}
body{background:#0d1117;color:#c9d1d9;font-family:'Cascadia Code','Fira Code','JetBrains Mono',ui-monospace,monospace;padding:1.5rem 2rem;font-size:13px;line-height:1.6}
pre{white-space:pre-wrap;word-break:break-word}
#data.htmx-request{opacity:.4;transition:opacity .3s .15s}
</style>
</head>
<body>
<div id="data" hx-get="/fragment" hx-trigger="every 5s" hx-swap="innerHTML">
<pre>%s</pre>
</div>
</body>
</html>`

func startWeb(addr string, doFetch func() (fetch.ChainSnapshot, fetch.EVMSnapshot, fetch.SystemSnapshot, fetch.DockerSnapshot)) {
	render := func() string {
		chain, ev, sys, docker := doFetch()
		var buf bytes.Buffer
		printAll(&buf, chain, ev, sys, docker)
		return html.EscapeString(buf.String())
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, page, render())
	})

	http.HandleFunc("/fragment", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, "<pre>%s</pre>", render())
	})

	go func() {
		log.Printf("web UI → http://localhost%s", addr)
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Printf("web server: %v", err)
		}
	}()
}
