package html

import (
	"fmt"
	"log"
	"net/http"

	"github.com/arkantos1482/cosmos-monitor/internal/fetch"
	"github.com/arkantos1482/cosmos-monitor/internal/model"
	"github.com/arkantos1482/cosmos-monitor/internal/report"
)

// FetchFunc loads live snapshots from the node.
type FetchFunc func() (fetch.ChainSnapshot, fetch.EVMSnapshot, fetch.SystemSnapshot, fetch.DockerSnapshot)

// Start serves the HTMX dashboard on addr (e.g. ":7777").
func Start(addr string, evmEndpoint string, doFetch FetchFunc) {
	render := func() (model.Report, string) {
		chain, ev, sys, docker := doFetch()
		d := report.Build(chain, ev, sys, docker, evmEndpoint)
		return d, RenderFragment(d)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		d, fragment := render()
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, FullPage(d.Moniker, fragment))
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
