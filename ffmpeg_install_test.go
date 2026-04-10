package core

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestDetectFFmpeg_ReturnsPath(t *testing.T) {
	info := DetectFFmpeg()
	if info.Path == "" {
		t.Error("expected non-empty Path")
	}
	// Found may be false in CI if ffmpeg missing; accept either.
	if info.Found && info.Version == "" {
		t.Error("Found=true requires Version")
	}
}

func TestInstallFFmpeg_EmitsProgress(t *testing.T) {
	body := make([]byte, 2048)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "2048")
		w.WriteHeader(200)
		w.Write(body)
	}))
	defer srv.Close()

	tmp := t.TempDir()
	dest := filepath.Join(tmp, "ffmpeg.bin")

	var lastTotal, lastDone int64
	err := downloadWithProgress(context.Background(), srv.URL, dest, func(done, total int64) {
		lastDone = done
		lastTotal = total
	})
	if err != nil {
		t.Fatalf("download: %v", err)
	}
	if lastTotal != 2048 {
		t.Errorf("total = %d", lastTotal)
	}
	if lastDone != 2048 {
		t.Errorf("done = %d", lastDone)
	}
	if stat, err := os.Stat(dest); err != nil || stat.Size() != 2048 {
		t.Errorf("dest: %v size=%v", err, stat)
	}
	_ = io.Discard
}
