package core

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestParseTimedText(t *testing.T) {
	xml := `<?xml version="1.0" encoding="utf-8" ?>
<transcript>
<text start="0.5" dur="2.0">Hello world</text>
<text start="2.5" dur="1.5">Second line</text>
</transcript>`
	plain, synced, err := parseTimedText(xml)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if !strings.Contains(plain, "Hello world") || !strings.Contains(plain, "Second line") {
		t.Errorf("plain missing lines: %q", plain)
	}
	if !strings.Contains(synced, "[00:00.50]") {
		t.Errorf("synced missing first timestamp: %q", synced)
	}
	if !strings.Contains(synced, "[00:02.50]") {
		t.Errorf("synced missing second timestamp: %q", synced)
	}
}

func TestFetchYouTubeCaptions_MockServer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("v") == "" {
			w.WriteHeader(400)
			return
		}
		w.Header().Set("Content-Type", "application/xml")
		w.Write([]byte(`<transcript><text start="0" dur="1">line1</text></transcript>`))
	}))
	defer srv.Close()

	// override endpoint via package variable
	prev := youtubeTimedTextURL
	youtubeTimedTextURL = srv.URL + "/api/timedtext"
	defer func() { youtubeTimedTextURL = prev }()

	res, err := FetchYouTubeCaptions("dQw4w9WgXcQ")
	if err != nil {
		t.Fatalf("FetchYouTubeCaptions: %v", err)
	}
	if res.PlainText == "" {
		t.Errorf("expected plain text, got empty")
	}
	if res.Source != "youtube-captions" {
		t.Errorf("source=%q want youtube-captions", res.Source)
	}
}

func TestFetchYouTubeCaptions_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}))
	defer srv.Close()

	prev := youtubeTimedTextURL
	youtubeTimedTextURL = srv.URL + "/api/timedtext"
	defer func() { youtubeTimedTextURL = prev }()

	_, err := FetchYouTubeCaptions("zzz")
	if err == nil {
		t.Errorf("expected error on 404")
	}
}
