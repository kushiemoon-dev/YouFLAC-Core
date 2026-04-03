package core

import "testing"

func TestIsQobuzURL_AlphanumericID(t *testing.T) {
	cases := []struct {
		url  string
		want bool
	}{
		{"https://www.qobuz.com/us-en/album/jacob-and-the-stone-emile-mosseri/qwc3banwgl2lb", true},
		{"https://www.qobuz.com/us-en/track/12345678", true},
		{"https://www.qobuz.com/us-en/album/some-album/abc123XYZ", true},
		{"https://www.youtube.com/watch?v=abc", false},
	}
	for _, tt := range cases {
		if got := IsQobuzURL(tt.url); got != tt.want {
			t.Errorf("IsQobuzURL(%q) = %v, want %v", tt.url, got, tt.want)
		}
	}
}
