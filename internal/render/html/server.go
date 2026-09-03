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

type dashServer struct {
	render   RenderFunc
	opts     panel.Options
	auth     Auth
	sessions *sessionStore
}

// NewHandler serves the dashboard, with an HTML login gate when Auth is set.
func NewHandler(render RenderFunc, opts panel.Options, auth Auth) http.Handler {
	s := &dashServer{
		render:   render,
		opts:     opts,
		auth:     auth,
		sessions: newSessionStore(),
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /login", s.getLogin)
	mux.HandleFunc("POST /login", s.postLogin)
	mux.HandleFunc("POST /logout", s.postLogout)
	mux.HandleFunc("/s/", s.serveSection)
	mux.HandleFunc("/delegate", s.serveDelegate)
	mux.HandleFunc("/", s.serveHome)
	return s.withLoginGate(mux)
}

func (s *dashServer) serveSection(w http.ResponseWriter, r *http.Request) {
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
	s.serveView(w, r, v)
}

func (s *dashServer) serveDelegate(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/delegate" {
		http.NotFound(w, r)
		return
	}
	s.serveView(w, r, panel.ViewDelegate)
}

func (s *dashServer) serveHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	s.serveView(w, r, panel.ViewHome)
}

func (s *dashServer) serveView(w http.ResponseWriter, r *http.Request, v panel.View) {
	d := s.render(v)
	fragment := RenderViewWithOptions(v, d, s.opts)
	status := panel.RenderStatusStrip(d)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if r.Header.Get("HX-Request") != "" && r.Header.Get("HX-Boosted") == "" {
		fmt.Fprint(w, panel.BuildStatusOOB(d)+fragment)
		return
	}
	fmt.Fprint(w, fullPage(pageMoniker(d), v, status, fragment, s.auth.enabled()))
}

// Start serves the dashboard on addr (e.g. ":7777").
func Start(addr string, evmEndpoint string, render RenderFunc, opts panel.Options, auth Auth) {
	if auth.enabled() {
		log.Printf("web UI → http://localhost%s (HTML login, all pages)", addr)
	} else {
		log.Printf("web UI → http://localhost%s (no login)", addr)
	}
	if err := http.ListenAndServe(addr, NewHandler(render, opts, auth)); err != nil {
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
