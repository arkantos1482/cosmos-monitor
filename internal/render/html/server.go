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

// Start serves the HTMX dashboard on addr (e.g. ":7777").
func Start(addr string, evmEndpoint string, render RenderFunc) {
	http.HandleFunc("/fragment", func(w http.ResponseWriter, r *http.Request) {
		v := panel.ParseView(r.URL.Query().Get("view"))
		d := render(v)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, WrapFragment(v, RenderView(v, d)))
	})

	http.HandleFunc("/s/", func(w http.ResponseWriter, r *http.Request) {
		slug := strings.TrimPrefix(r.URL.Path, "/s/")
		slug = strings.TrimSuffix(slug, "/")
		v := panel.ParseView(slug)
		if v == panel.ViewHome {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		d := render(v)
		fmt.Fprint(w, FullPage(pageMoniker(d), v, RenderView(v, d)))
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		v := panel.ViewHome
		d := render(v)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, FullPage(pageMoniker(d), v, RenderView(v, d)))
	})

	log.Printf("web UI → http://localhost%s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("web server: %v", err)
	}
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
