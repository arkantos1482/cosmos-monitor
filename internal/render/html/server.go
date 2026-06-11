package html

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/arkantos1482/cosmos-monitor/internal/fetchall"
	"github.com/arkantos1482/cosmos-monitor/internal/model"
	"github.com/arkantos1482/cosmos-monitor/internal/panel"
)

// RenderFunc builds a Report for the given dashboard view.
type RenderFunc func(v panel.View) model.Report

// Start serves the dashboard on addr (e.g. ":7777").
func Start(addr string, evmEndpoint string, render RenderFunc, opts panel.Options) {
	http.HandleFunc("/s/", func(w http.ResponseWriter, r *http.Request) {
		slug := strings.TrimPrefix(r.URL.Path, "/s/")
		slug = strings.TrimSuffix(slug, "/")
		v := panel.ParseView(slug)
		if v == panel.ViewHome {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		serveView(w, r, v, render, opts)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		serveView(w, r, panel.ViewHome, render, opts)
	})

	log.Printf("web UI → http://localhost%s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("web server: %v", err)
	}
}

func serveView(w http.ResponseWriter, r *http.Request, v panel.View, render RenderFunc, opts panel.Options) {
	d := render(v)
	fragment := RenderViewWithOptions(v, d, opts)
	status := panel.RenderStatusStrip(d)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// Poll-only HTMX (#data every 5s): OOB status + main fragment. Boost nav and direct loads: full page.
	if r.Header.Get("HX-Request") != "" && r.Header.Get("HX-Boosted") == "" {
		fmt.Fprint(w, panel.BuildStatusOOB(d)+fragment)
		return
	}
	fmt.Fprint(w, FullPage(pageMoniker(d), v, status, fragment))
}

func pageMoniker(d model.Report) string {
	if d.Moniker != "" {
		return d.Moniker
	}
	if m := fetchall.Moniker(); m != "" {
		return m
	}
	return "pmtop"
}
