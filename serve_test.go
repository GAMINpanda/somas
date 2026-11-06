package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func get(t *testing.T, h http.Handler, path string) (*http.Response, string) {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	res := rec.Result()
	body, _ := io.ReadAll(res.Body)
	return res, string(body)
}

func TestServesViewer(t *testing.T) {
	h := newViewerHandler([]byte(`{}`))
	res, body := get(t, h, "/")
	if res.StatusCode != http.StatusOK || !strings.Contains(body, "<title>AmoeBoid Patch Viewer</title>") {
		t.Error("GET / returned", res.StatusCode, "without the viewer page")
	}
}

func TestServesRun(t *testing.T) {
	run := `{"settings":{},"frames":[]}`
	h := newViewerHandler([]byte(run))
	res, body := get(t, h, "/run.json")
	if res.StatusCode != http.StatusOK || body != run {
		t.Error("GET /run.json returned", res.StatusCode, body)
	}
	if ct := res.Header.Get("Content-Type"); ct != "application/json" {
		t.Error("content type is", ct)
	}
	if cc := res.Header.Get("Cache-Control"); cc != "no-store" {
		t.Error("cache control is", cc, "expected no-store so the browser never shows an old run")
	}
}

func TestUnknownPathIsNotFound(t *testing.T) {
	res, _ := get(t, newViewerHandler(nil), "/nope.txt")
	if res.StatusCode != http.StatusNotFound {
		t.Error("GET /nope.txt returned", res.StatusCode, "expected 404")
	}
}
