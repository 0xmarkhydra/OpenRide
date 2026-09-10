package httpapi

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestCanonicalRequestHashIgnoresJSONObjectFieldOrder(t *testing.T) {
	r1 := &http.Request{Method: http.MethodPost, URL: mustURL(t, "/v2/requests"), Body: io.NopCloser(strings.NewReader(`{"b":2,"a":1}`))}
	r2 := &http.Request{Method: http.MethodPost, URL: mustURL(t, "/v2/requests"), Body: io.NopCloser(strings.NewReader(`{"a":1,"b":2}`))}
	h1, err := canonicalRequestHash(r1)
	if err != nil { t.Fatal(err) }
	h2, err := canonicalRequestHash(r2)
	if err != nil { t.Fatal(err) }
	if h1 != h2 { t.Fatalf("expected canonical hashes to match: %s != %s", h1, h2) }
}

func TestCanonicalRequestHashIncludesPath(t *testing.T) {
	r1 := &http.Request{Method: http.MethodPost, URL: mustURL(t, "/v2/requests/req_a/cancel"), Body: io.NopCloser(strings.NewReader(`{}`))}
	r2 := &http.Request{Method: http.MethodPost, URL: mustURL(t, "/v2/requests/req_b/cancel"), Body: io.NopCloser(strings.NewReader(`{}`))}
	h1, err := canonicalRequestHash(r1)
	if err != nil { t.Fatal(err) }
	h2, err := canonicalRequestHash(r2)
	if err != nil { t.Fatal(err) }
	if h1 == h2 { t.Fatal("expected different paths to produce different hashes") }
}

func mustURL(t *testing.T, path string) *http.URL {
	t.Helper()
	panic("replaced by net/url helper")
}
