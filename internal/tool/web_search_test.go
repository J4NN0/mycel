package tool

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/J4NN0/mycel/internal/logger"
)

func TestSearchResults(t *testing.T) {
	var gotQuery, gotFormat string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == searchHealthPath {
			return
		}
		gotQuery, gotFormat = r.URL.Query().Get("q"), r.URL.Query().Get("format")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"results":[
			{"title":"Go 1.26 release notes","url":"https://go.dev/doc/go1.26","content":"  Go 1.26   adds  something.  "},
			{"title":"Second","url":"https://example.org/2","content":""}
		]}`))
	}))
	defer srv.Close()

	out, err := execSearch(t, srv.URL, "  go 1.26 release  ")
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if gotQuery != "go 1.26 release" {
		t.Errorf("query sent as %q", gotQuery)
	}
	if gotFormat != "json" {
		t.Errorf("format sent as %q, want json", gotFormat)
	}
	if !strings.Contains(out, "Go 1.26 adds something.") {
		t.Errorf("snippet whitespace not normalized:\n%s", out)
	}
	if !strings.Contains(out, "https://go.dev/doc/go1.26") {
		t.Errorf("result URL missing:\n%s", out)
	}
	if !strings.Contains(out, fetchName) {
		t.Errorf("results do not point at %s for the full page:\n%s", fetchName, out)
	}
}

func TestSearchNoResults(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == searchHealthPath {
			return
		}
		w.Write([]byte(`{"results":[]}`))
	}))
	defer srv.Close()

	out, err := execSearch(t, srv.URL, "nothing at all")
	if err != nil {
		t.Fatalf("an empty result set should not be an error: %v", err)
	}
	if !strings.Contains(out, "No results") {
		t.Errorf("empty result set not reported plainly: %q", out)
	}
}

// SearXNG answers with 403 when the json output format is not enabled in its settings, which is the
// default. The error has to name that, or the failure looks like the instance being down.
func TestSearchWhenJSONFormatDisabled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == searchHealthPath {
			return
		}
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("<html>Forbidden</html>"))
	}))
	defer srv.Close()

	_, err := execSearch(t, srv.URL, "anything")
	if err == nil {
		t.Fatal("expected an error on 403")
	}
	if !strings.Contains(err.Error(), "json") {
		t.Errorf("403 error does not mention the json format: %v", err)
	}
}

func TestSearchNotRegisteredWithoutValidURL(t *testing.T) {
	for _, addr := range []string{"", "localhost:8888", "not a url"} {
		if s := NewWebSearch(logger.New("test", "panic"), addr); s != nil {
			t.Errorf("search tool registered with SEARXNG_URL %q", addr)
		}
	}
}

func TestSearchNotRegisteredWhenUnreachable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	url := srv.URL
	srv.Close()

	if s := NewWebSearch(logger.New("test", "panic"), url); s != nil {
		t.Error("search tool registered against an instance that is not running")
	}
}

func TestSearchNotRegisteredWhenUnhealthy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	if s := NewWebSearch(logger.New("test", "panic"), srv.URL); s != nil {
		t.Error("search tool registered against a host that is not SearXNG")
	}
}

func execSearch(t *testing.T, searxngURL, query string) (string, error) {
	t.Helper()

	s := NewWebSearch(logger.New("test", "error"), searxngURL)
	if s == nil {
		t.Fatalf("NewSearch returned nil for %q", searxngURL)
	}

	args, err := json.Marshal(searchArgs{Query: query})
	if err != nil {
		t.Fatalf("marshal args: %v", err)
	}

	return s.Execute(context.Background(), args)
}
