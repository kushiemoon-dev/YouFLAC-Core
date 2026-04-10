package core

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetThumbnailMax(t *testing.T) {
	tests := []struct {
		name       string
		available  map[string]int
		wantSuffix string
	}{
		{
			name: "maxres available",
			available: map[string]int{
				"/vi/xxx/maxresdefault.jpg": 200,
				"/vi/xxx/sddefault.jpg":     200,
				"/vi/xxx/hqdefault.jpg":     200,
			},
			wantSuffix: "maxresdefault.jpg",
		},
		{
			name: "fallback to sd",
			available: map[string]int{
				"/vi/xxx/maxresdefault.jpg": 404,
				"/vi/xxx/sddefault.jpg":     200,
				"/vi/xxx/hqdefault.jpg":     200,
			},
			wantSuffix: "sddefault.jpg",
		},
		{
			name: "fallback to hq",
			available: map[string]int{
				"/vi/xxx/maxresdefault.jpg": 404,
				"/vi/xxx/sddefault.jpg":     404,
				"/vi/xxx/hqdefault.jpg":     200,
			},
			wantSuffix: "hqdefault.jpg",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if code, ok := tt.available[r.URL.Path]; ok {
					w.WriteHeader(code)
					return
				}
				w.WriteHeader(404)
			}))
			defer srv.Close()

			prev := youtubeThumbnailBase
			youtubeThumbnailBase = srv.URL + "/vi"
			defer func() { youtubeThumbnailBase = prev }()

			got := GetThumbnailMax("xxx")
			if !strings.HasSuffix(got, tt.wantSuffix) {
				t.Errorf("got=%q want suffix=%q", got, tt.wantSuffix)
			}
		})
	}
}
