package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBulkFetchLyrics_WalksAudioFiles(t *testing.T) {
	dir := t.TempDir()
	files := []string{"track1.flac", "track2.mka", "track3.mkv", "notes.txt"}
	for _, f := range files {
		if err := os.WriteFile(filepath.Join(dir, f), []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	origResolver := bulkLyricsResolver
	bulkLyricsResolver = func(artist, title, album, videoID string) (*LyricsResult, error) {
		return &LyricsResult{PlainText: "la la la", Source: "stub"}, nil
	}
	defer func() { bulkLyricsResolver = origResolver }()

	results, err := BulkFetchLyrics(dir)
	if err != nil {
		t.Fatalf("BulkFetchLyrics: %v", err)
	}
	if len(results) != 3 {
		t.Errorf("expected 3 audio files, got %d", len(results))
	}
}
