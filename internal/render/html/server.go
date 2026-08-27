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

// securityCSP allows first-party JS, inlined theme CSS, and Google Fonts.
// Delegate talks to the chain through the injected wallet, not page fetch.
const securityCSP = "default-src 'none'; script-src 'self'; style-src 'unsafe-inline' https://fonts.googleapis.com; font-src https://fonts.gstatic.com; img-src 'self' data:; connect-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'; object-src 'none'"

func writeSecurityHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Security-Policy", securityCSP)
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Referrer-Policy", "no-referrer")
}

func serveStaticJS(body string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		writeSecurityHeaders(w)
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		if r.Method == http.MethodHead {
			return
		}
		_, _ = w.Write([]byte(body))
	}
}

func registerRoutes(mux *http.ServeMux, render RenderFunc, opts panel.Options) {
	mux.HandleFunc("/static/htmx.min.js", serveStaticJS(htmxJS))
	mux.HandleFunc("/static/htmx-init.js", serveStaticJS(htmxInitJS))
	mux.HandleFunc("/static/ethers.umd.min.js", serveStaticJS(ethersJS))
	mux.HandleFunc("/static/delegate.js", serveStaticJS(delegateJS))

	mux.HandleFunc("/s/", func(w http.ResponseWriter, r *http.Request) {
		slug := strings.TrimPrefix(r.URL.Path, "/s/")
		slug = strings.TrimSuffix(slug, "/")
		v := panel.ParseView(slug)
		if v == panel.ViewHome {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		if v == panel.ViewDelegate {
			http.Redirect(w, r, "/delegate", http.StatusFound)
			return
		}
		serveView(w, r, v, render, opts)
	})

	mux.HandleFunc("/delegate", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/delegate" {
			http.NotFound(w, r)
			return
		}
		serveView(w, r, panel.ViewDelegate, render, opts)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		serveView(w, r, panel.ViewHome, render, opts)
	})
}

// Start serves the dashboard on addr (e.g. ":7777").
func Start(addr string, evmEndpoint string, render RenderFunc, opts panel.Options) {
	mux := http.NewServeMux()
	registerRoutes(mux, render, opts)
	log.Printf("web UI → http://localhost%s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("web server: %v", err)
	}
}

func serveView(w http.ResponseWriter, r *http.Request, v panel.View, render RenderFunc, opts panel.Options) {
	d := render(v)
	fragment := RenderViewWithOptions(v, d, opts)
	status := panel.RenderStatusStrip(d)
	writeSecurityHeaders(w)
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
