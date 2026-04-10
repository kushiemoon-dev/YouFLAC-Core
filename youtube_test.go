package core

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestParseChannelAssetsJSON(t *testing.T) {
	jsonBody := `{
		"uploader_id": "UCabc",
		"channel_url": "https://youtube.com/channel/UCabc",
		"channel": "Test Channel",
		"thumbnails": [
			{"id": "avatar_uncropped", "url": "https://yt/avatar.jpg", "preference": 1},
			{"id": "banner_uncropped", "url": "https://yt/banner.jpg", "preference": 2}
		]
	}`
	assets, err := parseChannelAssetsJSON([]byte(jsonBody))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if assets.ChannelID != "UCabc" {
		t.Errorf("channelID=%q", assets.ChannelID)
	}
	if assets.ChannelName != "Test Channel" {
		t.Errorf("name=%q", assets.ChannelName)
	}
	if assets.AvatarURL != "https://yt/avatar.jpg" {
		t.Errorf("avatar=%q", assets.AvatarURL)
	}
	if assets.BannerURL != "https://yt/banner.jpg" {
		t.Errorf("banner=%q", assets.BannerURL)
	}
}

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
