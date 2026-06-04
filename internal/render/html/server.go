package html

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/arkantos1482/cosmos-monitor/internal/fetch"
	"github.com/arkantos1482/cosmos-monitor/internal/model"
	"github.com/arkantos1482/cosmos-monitor/internal/panel"
	"github.com/arkantos1482/cosmos-monitor/internal/report"
)

// FetchFunc loads live snapshots from the node.
type FetchFunc func() (fetch.ChainSnapshot, fetch.EVMSnapshot, fetch.SystemSnapshot, fetch.DockerSnapshot)

// Start serves the HTMX dashboard on addr (e.g. ":7777").
func Start(addr string, evmEndpoint string, doFetch FetchFunc) {
	render := func() model.Report {
		chain, ev, sys, docker := doFetch()
		return report.Build(chain, ev, sys, docker, evmEndpoint)
	}

	http.HandleFunc("/fragment", func(w http.ResponseWriter, r *http.Request) {
		d := render()
		v := panel.ParseView(r.URL.Query().Get("view"))
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, RenderView(v, d))
	})

	http.HandleFunc("/s/", func(w http.ResponseWriter, r *http.Request) {
		d := render()
		slug := strings.TrimPrefix(r.URL.Path, "/s/")
		slug = strings.TrimSuffix(slug, "/")
		v := panel.ParseView(slug)
		if v == panel.ViewHome {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, FullPage(d.Moniker, v, RenderView(v, d)))
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		d := render()
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, FullPage(d.Moniker, panel.ViewHome, RenderView(panel.ViewHome, d)))
	})

	log.Printf("web UI → http://localhost%s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("web server: %v", err)
	}
}
