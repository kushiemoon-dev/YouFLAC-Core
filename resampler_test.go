package core

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func generateSineWave(t *testing.T, path string, sampleRate, bitDepth int, seconds float64) {
	t.Helper()
	if err := CheckFFmpegInstalled(); err != nil {
		t.Skipf("ffmpeg not installed: %v", err)
	}
	codec := "pcm_s16le"
	sampleFmt := "s16"
	switch bitDepth {
	case 24:
		codec, sampleFmt = "pcm_s24le", "s32"
	case 32:
		codec, sampleFmt = "pcm_s32le", "s32"
	}
	args := []string{
		"-y", "-f", "lavfi",
		"-i", fmt.Sprintf("sine=frequency=1000:sample_rate=%d:duration=%.3f", sampleRate, seconds),
		"-sample_fmt", sampleFmt, "-c:a", codec, path,
	}
	cmd := exec.Command(GetFFmpegPath(), args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("generateSineWave failed: %v - %s", err, stderr.String())
	}
	if st, err := os.Stat(path); err != nil || st.Size() == 0 {
		t.Fatalf("generateSineWave produced no output: %v", err)
	}
}

func TestResample_ValidateMissingInput(t *testing.T) {
	err := Resample(context.Background(), ResampleOptions{
		InputPath:  "/nonexistent/file.wav",
		OutputPath: filepath.Join(t.TempDir(), "out.flac"),
		SampleRate: 48000,
		BitDepth:   16,
		Format:     "flac",
	})
	if err == nil {
		t.Fatal("expected error for missing input, got nil")
	}
}

func TestResample_ValidateUnsupportedRate(t *testing.T) {
	in := filepath.Join(t.TempDir(), "in.wav")
	generateSineWave(t, in, 44100, 16, 0.5)
	err := Resample(context.Background(), ResampleOptions{
		InputPath:  in,
		OutputPath: filepath.Join(t.TempDir(), "out.flac"),
		SampleRate: 12345,
		BitDepth:   16,
		Format:     "flac",
	})
	if err == nil {
		t.Fatal("expected validation error for unsupported sample rate")
	}
}

func TestResample_Upsample44kTo96k(t *testing.T) {
	if err := CheckFFmpegInstalled(); err != nil {
		t.Skipf("ffmpeg not installed: %v", err)
	}
	tmp := t.TempDir()
	in := filepath.Join(tmp, "in.wav")
	out := filepath.Join(tmp, "out.flac")
	generateSineWave(t, in, 44100, 16, 1.0)
	if err := Resample(context.Background(), ResampleOptions{
		InputPath:  in,
		OutputPath: out,
		SampleRate: 96000,
		BitDepth:   24,
		Format:     "flac",
	}); err != nil {
		t.Fatalf("Resample failed: %v", err)
	}
	info, err := AnalyzeAudio(out)
	if err != nil {
		t.Fatalf("AnalyzeAudio failed: %v", err)
	}
	if info.SampleRate != 96000 {
		t.Errorf("want 96000 Hz, got %d", info.SampleRate)
	}
	if info.BitsPerSample != 24 {
		t.Errorf("want 24-bit, got %d", info.BitsPerSample)
	}
}

func TestResample_Downsample96kTo44kWithDither(t *testing.T) {
	if err := CheckFFmpegInstalled(); err != nil {
		t.Skipf("ffmpeg not installed: %v", err)
	}
	tmp := t.TempDir()
	in := filepath.Join(tmp, "in.wav")
	out := filepath.Join(tmp, "out.wav")
	generateSineWave(t, in, 96000, 24, 1.0)
	if err := Resample(context.Background(), ResampleOptions{
		InputPath:  in,
		OutputPath: out,
		SampleRate: 44100,
		BitDepth:   16,
		Dither:     true,
		Format:     "wav",
	}); err != nil {
		t.Fatalf("Resample failed: %v", err)
	}
	info, err := AnalyzeAudio(out)
	if err != nil {
		t.Fatalf("AnalyzeAudio failed: %v", err)
	}
	if info.SampleRate != 44100 {
		t.Errorf("want 44100 Hz, got %d", info.SampleRate)
	}
}
