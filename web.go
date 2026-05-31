package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"html"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/arkantos1482/cosmos-monitor/fetch"
)

//go:embed web.html
var pageTemplate string

var (
	ansiRe      = regexp.MustCompile(`\x1b\[(\d+)m`)
	ansiClasses = map[string]string{
		"1": "b", "2": "dim",
		"31": "r", "32": "g", "33": "y", "36": "c", "97": "w",
	}
)

func ansiToHTML(s string) string {
	var b strings.Builder
	open, last := 0, 0
	for _, m := range ansiRe.FindAllStringSubmatchIndex(s, -1) {
		b.WriteString(html.EscapeString(s[last:m[0]]))
		last = m[1]
		code := s[m[2]:m[3]]
		if code == "0" {
			for open > 0 {
				b.WriteString("</span>")
				open--
			}
		} else if cls, ok := ansiClasses[code]; ok {
			b.WriteString(`<span class="` + cls + `">`)
			open++
		}
	}
	b.WriteString(html.EscapeString(s[last:]))
	for open > 0 {
		b.WriteString("</span>")
		open--
	}
	return b.String()
}

func startWeb(addr string, doFetch func() (fetch.ChainSnapshot, fetch.EVMSnapshot, fetch.SystemSnapshot, fetch.DockerSnapshot)) {
	render := func() string {
		chain, ev, sys, docker := doFetch()
		var buf bytes.Buffer
		printAll(&buf, chain, ev, sys, docker)
		return ansiToHTML(buf.String())
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, strings.ReplaceAll(pageTemplate, "{{CONTENT}}", render()))
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
