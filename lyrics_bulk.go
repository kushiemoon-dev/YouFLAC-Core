package core

import (
	"os"
	"path/filepath"
	"strings"
)

// bulkLyricsResolver is overridable in tests.
var bulkLyricsResolver = func(artist, title, album, videoID string) (*LyricsResult, error) {
	return FetchLyricsWithFallback(artist, title, album, videoID)
}

var bulkAudioExts = map[string]bool{
	".flac": true, ".mka": true, ".mkv": true, ".mp3": true, ".m4a": true,
	".opus": true, ".ogg": true, ".wav": true,
}

// BulkFetchLyrics walks dir and fetches+writes .lrc for every audio file.
// Returns a map of filepath → error (nil = success).
func BulkFetchLyrics(dir string) (map[string]error, error) {
	results := make(map[string]error)
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, werr error) error {
		if werr != nil {
			return werr
		}
		if d.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if !bulkAudioExts[ext] {
			return nil
		}
		base := strings.TrimSuffix(filepath.Base(path), ext)
		artist, title := "", base
		if i := strings.Index(base, " - "); i > 0 {
			artist = base[:i]
			title = base[i+3:]
		}
		res, ferr := bulkLyricsResolver(artist, title, "", "")
		if ferr != nil {
			results[path] = ferr
			return nil
		}
		lrc := res.SyncedLyrics
		if lrc == "" {
			lrc = res.PlainText
		}
		lrcPath := strings.TrimSuffix(path, ext) + ".lrc"
		results[path] = os.WriteFile(lrcPath, []byte(lrc), 0644)
		return nil
	})
	return results, err
}
