package main

import (
	_ "embed"
	"html/template"
	"log"
	"net/http"

	"github.com/arkantos1482/cosmos-monitor/fetch"
)

//go:embed dashboard.html
var dashboardHTML string

var dashTmpl = template.Must(
	template.New("dashboard").Funcs(template.FuncMap{
		"mmColor": meterColor,
		"not":     func(b bool) bool { return !b },
	}).Parse(dashboardHTML),
)

func startWeb(addr string, doFetch func() (fetch.ChainSnapshot, fetch.EVMSnapshot, fetch.SystemSnapshot, fetch.DockerSnapshot)) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		chain, ev, sys, docker := doFetch()
		data := buildWebData(chain, ev, sys, docker)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := dashTmpl.ExecuteTemplate(w, "page", data); err != nil {
			log.Printf("template error: %v", err)
		}
	})

	http.HandleFunc("/fragment", func(w http.ResponseWriter, r *http.Request) {
		chain, ev, sys, docker := doFetch()
		data := buildWebData(chain, ev, sys, docker)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := dashTmpl.ExecuteTemplate(w, "fragment", data); err != nil {
			log.Printf("template error: %v", err)
		}
	})

	go func() {
		log.Printf("web UI → http://localhost%s", addr)
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Printf("web server: %v", err)
		}
	}()
}
