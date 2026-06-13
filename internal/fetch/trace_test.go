package fetch

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDoJSONRecordsExchange(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	BeginTrace()
	var out struct {
		OK bool `json:"ok"`
	}
	if err := doJSON(srv.URL, &out); err != nil {
		t.Fatal(err)
	}
	exchanges := EndTrace()

	if len(exchanges) != 1 {
		t.Fatalf("expected 1 exchange, got %d", len(exchanges))
	}
	e := exchanges[0]
	if e.Method != "GET" || !e.OK || e.Request != "(none)" {
		t.Fatalf("unexpected exchange: %+v", e)
	}
	if e.Response != `{"ok":true}` {
		t.Fatalf("unexpected response: %q", e.Response)
	}
}
