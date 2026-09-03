package html

import (
	"bytes"
	_ "embed"
	"html/template"
	"net/http"
)

//go:embed templates/login.html
var loginHTML string

var loginTmpl = template.Must(template.New("login").Parse(loginHTML))

type loginData struct {
	CSS   template.CSS
	Next  string
	Error string
}

func writeLoginPage(w http.ResponseWriter, status int, next, errMsg string) {
	var buf bytes.Buffer
	_ = loginTmpl.Execute(&buf, loginData{
		CSS:   template.CSS(themeCSS),
		Next:  next,
		Error: errMsg,
	})
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_, _ = w.Write(buf.Bytes())
}

func (s *dashServer) getLogin(w http.ResponseWriter, r *http.Request) {
	next := formNext(r)
	if s.sessions.valid(sessionID(r)) {
		http.Redirect(w, r, next, http.StatusSeeOther)
		return
	}
	writeLoginPage(w, http.StatusOK, next, "")
}

func (s *dashServer) postLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeLoginPage(w, http.StatusBadRequest, "/", "Could not read the form.")
		return
	}
	next := safeNext(r.Form.Get("next"))
	user := r.Form.Get("username")
	pass := r.Form.Get("password")
	if !credsOK(s.auth, user, pass) {
		writeLoginPage(w, http.StatusUnauthorized, next, "Wrong user or password.")
		return
	}
	id := s.sessions.issue()
	setSessionCookie(w, r, id, int(sessionTTL.Seconds()))
	http.Redirect(w, r, next, http.StatusSeeOther)
}

func (s *dashServer) postLogout(w http.ResponseWriter, r *http.Request) {
	s.sessions.revoke(sessionID(r))
	setSessionCookie(w, r, "", -1)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
