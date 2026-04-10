package core

import (
	"strings"
	"testing"
)

func TestGetMKVMetadataArgs_Phase3Tags(t *testing.T) {
	m := &Metadata{
		Title:      "Song",
		YouTubeURL: "https://youtu.be/abc",
		ViewCount:  42,
		UploadDate: "20240115",
	}
	args := GetMKVMetadataArgs(m)
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "YOUTUBE_URL=https://youtu.be/abc") {
		t.Errorf("missing YOUTUBE_URL: %v", args)
	}
	if !strings.Contains(joined, "VIEW_COUNT=42") {
		t.Errorf("missing VIEW_COUNT: %v", args)
	}
	if !strings.Contains(joined, "date=2024-01-15") {
		t.Errorf("missing ISO date: %v", args)
	}
}

func TestDetectExplicit(t *testing.T) {
	tests := []struct {
		title string
		want  bool
	}{
		{"Some Song [Explicit]", true},
		{"Some Song (Explicit)", true},
		{"Some Song - Explicit Version", true},
		{"Some Song [explicit]", true},
		{"Explicitly Yours", false},
		{"Clean Version", false},
		{"", false},
	}
	for _, tc := range tests {
		t.Run(tc.title, func(t *testing.T) {
			if got := DetectExplicit(tc.title); got != tc.want {
				t.Errorf("DetectExplicit(%q) = %v, want %v", tc.title, got, tc.want)
			}
		})
	}
}
