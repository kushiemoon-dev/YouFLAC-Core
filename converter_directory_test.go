package core

import (
	"context"
	"os/exec"
	"path/filepath"
	"testing"
)

// makeTempWAVForDir creates a minimal silent WAV file via ffmpeg.
// Skips the test if ffmpeg is not available.
func makeTempWAVForDir(t *testing.T, dir, name string) string {
	t.Helper()
	out := filepath.Join(dir, name)
	cmd := exec.Command("ffmpeg", "-y", "-f", "lavfi",
		"-i", "sine=frequency=440:sample_rate=44100:duration=0.1",
		"-c:a", "pcm_s16le", out)
	if err := cmd.Run(); err != nil {
		t.Skipf("ffmpeg not available: %v", err)
	}
	return out
}

func TestConvertDirectory_InvalidDir(t *testing.T) {
	opts := ConvertDirOptions{
		Dir:          "/nonexistent/path/does/not/exist",
		TargetFormat: "mp3",
	}
	err := ConvertDirectory(context.Background(), opts, func(r DirConvertResult) {})
	if err == nil {
		t.Fatal("expected error for non-existent directory, got nil")
	}
}

func TestConvertDirectory_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	opts := ConvertDirOptions{
		Dir:          dir,
		TargetFormat: "mp3",
	}

	var results []DirConvertResult
	err := ConvertDirectory(context.Background(), opts, func(r DirConvertResult) {
		results = append(results, r)
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result (final summary), got %d", len(results))
	}
	final := results[0]
	if !final.Done {
		t.Error("expected Done=true on final item")
	}
	if final.Total != 0 {
		t.Errorf("expected Total=0, got %d", final.Total)
	}
}

func TestConvertDirectory_SingleFile(t *testing.T) {
	dir := t.TempDir()
	makeTempWAVForDir(t, dir, "track.wav")

	opts := ConvertDirOptions{
		Dir:          dir,
		TargetFormat: "mp3",
	}

	var results []DirConvertResult
	err := ConvertDirectory(context.Background(), opts, func(r DirConvertResult) {
		results = append(results, r)
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Expect 1 file result + 1 final summary
	if len(results) < 2 {
		t.Fatalf("expected at least 2 results, got %d", len(results))
	}

	fileResult := results[0]
	if fileResult.Done {
		t.Error("first result should not be Done")
	}
	if fileResult.SourcePath == "" {
		t.Error("expected SourcePath to be set on file result")
	}

	final := results[len(results)-1]
	if !final.Done {
		t.Error("last result should have Done=true")
	}
	if final.Total != 1 {
		t.Errorf("expected Total=1, got %d", final.Total)
	}
	if final.Succeeded != 1 {
		t.Errorf("expected Succeeded=1, got %d", final.Succeeded)
	}
	if final.Failed != 0 {
		t.Errorf("expected Failed=0, got %d", final.Failed)
	}
}

func TestConvertDirectory_ContextCancel(t *testing.T) {
	dir := t.TempDir()
	makeTempWAVForDir(t, dir, "a.wav")
	makeTempWAVForDir(t, dir, "b.wav")

	ctx, cancel := context.WithCancel(context.Background())

	var count int
	err := ConvertDirectory(ctx, ConvertDirOptions{Dir: dir, TargetFormat: "mp3"}, func(r DirConvertResult) {
		count++
		// Cancel after the first file result
		if !r.Done {
			cancel()
		}
	})

	if err == nil {
		t.Error("expected context cancellation error, got nil")
	}
	// cancel() is called synchronously in the callback after the first file.
	// ConvertDirectory checks ctx.Done() before each file, so only 1 result is expected.
	// count >= 2 means the second file was also processed — early exit did not work.
	if count >= 2 {
		t.Errorf("expected early exit after first file, but got %d results", count)
	}
}
