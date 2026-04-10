package core

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestAnalyzeAudio_ExtendedFormats(t *testing.T) {
	if err := CheckFFmpegInstalled(); err != nil {
		t.Skipf("ffmpeg not installed: %v", err)
	}

	tmpDir := t.TempDir()

	tests := []struct {
		ext       string
		codec     string
		extraArgs []string
	}{
		{
			ext:       "flac",
			codec:     "flac",
			extraArgs: []string{"-c:a", "flac"},
		},
		{
			ext:       "mp3",
			codec:     "mp3",
			extraArgs: []string{"-c:a", "libmp3lame", "-q:a", "2"},
		},
		{
			ext:       "m4a",
			codec:     "m4a",
			extraArgs: []string{"-c:a", "aac", "-b:a", "128k"},
		},
		{
			ext:       "ogg",
			codec:     "ogg",
			extraArgs: []string{"-c:a", "libvorbis", "-q:a", "4"},
		},
		{
			ext:       "wav",
			codec:     "wav",
			extraArgs: []string{"-c:a", "pcm_s16le"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.ext, func(t *testing.T) {
			outPath := filepath.Join(tmpDir, "test_"+tt.ext+"."+tt.ext)

			args := []string{
				"-y", "-f", "lavfi",
				"-i", "sine=frequency=440:sample_rate=44100:duration=0.2",
			}
			args = append(args, tt.extraArgs...)
			args = append(args, outPath)

			cmd := exec.Command(GetFFmpegPath(), args...)
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			if err := cmd.Run(); err != nil {
				t.Skipf("ffmpeg could not generate %s: %v - %s", tt.ext, err, stderr.String())
			}

			analysis, err := AnalyzeAudio(outPath)
			if err != nil {
				t.Fatalf("AnalyzeAudio(%s) error: %v", tt.ext, err)
			}
			if analysis.SampleRate <= 0 {
				t.Errorf("SampleRate = %d, want > 0", analysis.SampleRate)
			}
			if analysis.Codec == "" {
				t.Error("Codec is empty, want non-empty")
			}
			t.Logf("%s: codec=%s sampleRate=%d bitDepth=%d", tt.ext, analysis.Codec, analysis.SampleRate, analysis.BitsPerSample)
		})
	}
}
