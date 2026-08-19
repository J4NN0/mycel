package tool

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/J4NN0/mycel/internal/logger"
)

func TestReadableURL(t *testing.T) {
	for _, bad := range []string{"", "example.com/page", "ftp://example.com", "file:///etc/passwd", "not a url"} {
		got, err := readableURL(bad)
		if err == nil {
			t.Errorf("readableURL(%q) accepted it as %q", bad, got)
		}
	}

	_, err := readableURL(" https://example.com/page ")
	if err != nil {
		t.Errorf("readableURL rejected a valid URL: %v", err)
	}
}

// A local test server stands in for Ollama or Redis: reachable from this machine, and exactly what a
// fetch must not be able to reach.
func TestFetchRefusesLocalAddresses(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("<html><title>secret</title><body>local service</body></html>"))
	}))
	defer srv.Close()

	f := NewFetchURL(logger.New("test", "error"))
	args, err := json.Marshal(fetchArgs{URL: srv.URL})
	if err != nil {
		t.Fatalf("marshal args: %v", err)
	}

	out, err := f.Execute(context.Background(), args)
	if err == nil {
		t.Fatalf("fetch reached a local address and returned:\n%s", out)
	}
	if !strings.Contains(err.Error(), "local or private network") {
		t.Errorf("unexpected refusal reason: %v", err)
	}
}

func TestIsPublicIP(t *testing.T) {
	private := []string{"127.0.0.1", "::1", "10.0.0.5", "192.168.1.10", "172.16.0.1", "169.254.169.254", "0.0.0.0", "224.0.0.1"}
	for _, addr := range private {
		if isPublicIP(net.ParseIP(addr)) {
			t.Errorf("%s treated as public", addr)
		}
	}

	for _, addr := range []string{"1.1.1.1", "93.184.216.34", "2606:2800:220:1:248:1893:25c8:1946"} {
		if !isPublicIP(net.ParseIP(addr)) {
			t.Errorf("%s treated as private", addr)
		}
	}
}

func TestHTMLToText(t *testing.T) {
	const doc = `<html><head><title>  Hello World </title>
	<style>body{color:red}</style><script>var x = "not text";</script></head>
	<body><nav>Home About</nav><h1>Heading</h1><p>First   paragraph
	spanning lines.</p><p>Second &amp; last.</p>
	<ul><li>one</li><li>two</li></ul>
	<noscript>enable js</noscript></body></html>`

	title, text, err := htmlToText(strings.NewReader(doc))
	if err != nil {
		t.Fatalf("htmlToText: %v", err)
	}

	if title != "Hello World" {
		t.Errorf("title = %q, want %q", title, "Hello World")
	}
	for _, unwanted := range []string{"color:red", "not text", "enable js"} {
		if strings.Contains(text, unwanted) {
			t.Errorf("text leaked %q:\n%s", unwanted, text)
		}
	}
	if !strings.Contains(text, "Second & last.") {
		t.Errorf("entities not decoded:\n%s", text)
	}
	if !strings.Contains(text, "First paragraph spanning lines.") {
		t.Errorf("whitespace inside a paragraph not collapsed:\n%s", text)
	}
}

func TestTruncate(t *testing.T) {
	if got := truncate("short", 10); got != "short" {
		t.Errorf("truncate padded or cut a short string: %q", got)
	}

	got := truncate(strings.Repeat("x", 20), 10)
	if !strings.HasPrefix(got, strings.Repeat("x", 10)) || strings.Contains(got, strings.Repeat("x", 11)) {
		t.Errorf("truncate cut at the wrong place: %q", got)
	}
	if !strings.Contains(got, "truncated") {
		t.Errorf("truncate cut silently: %q", got)
	}
}
