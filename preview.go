package core

// PreviewAudio is a pure integration of yt-dlp + ffmpeg.
// There is no business logic to test in isolation — skip unit tests for this file.
// Integration tests would require both binaries present and network access.

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// PreviewAudio returns a ReadCloser streaming up to maxSeconds seconds of audio
// in OGG/Vorbis format, suitable for browser playback via HTML5 <audio>.
// The caller must close the returned ReadCloser to release resources.
func PreviewAudio(ctx context.Context, videoURL string, maxSeconds int) (io.ReadCloser, error) {
	if _, err := exec.LookPath("yt-dlp"); err != nil {
		return nil, fmt.Errorf("yt-dlp not found: install yt-dlp to enable audio preview")
	}
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return nil, fmt.Errorf("ffmpeg not found: install ffmpeg to enable audio preview")
	}

	// SSRF guard: only allow YouTube URLs
	if !strings.HasPrefix(videoURL, "https://www.youtube.com/") && !strings.HasPrefix(videoURL, "https://youtu.be/") {
		return nil, fmt.Errorf("unsupported URL: only YouTube URLs are accepted")
	}

	// Get direct audio stream URL from yt-dlp
	ytCmd := exec.CommandContext(ctx, "yt-dlp",
		"--get-url",
		"-f", "bestaudio[ext=webm]/bestaudio",
		"--no-playlist",
		videoURL,
	)
	out, err := ytCmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			msg = err.Error()
		}
		return nil, fmt.Errorf("yt-dlp failed: %s", msg)
	}

	audioURL := strings.SplitN(strings.TrimSpace(string(out)), "\n", 2)[0]
	if audioURL == "" {
		return nil, fmt.Errorf("yt-dlp returned empty URL for %s", videoURL)
	}

	// Stream through ffmpeg, encode to OGG/Vorbis for browser compatibility
	ffCmd := exec.CommandContext(ctx, "ffmpeg",
		"-i", audioURL,
		"-t", fmt.Sprintf("%d", maxSeconds),
		"-acodec", "libvorbis",
		"-f", "ogg",
		"-loglevel", "error",
		"pipe:1",
	)

	stdout, err := ffCmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("ffmpeg pipe setup failed: %w", err)
	}

	if err := ffCmd.Start(); err != nil {
		return nil, fmt.Errorf("ffmpeg start failed: %w", err)
	}

	return &previewReadCloser{ReadCloser: stdout, cmd: ffCmd}, nil
}

// previewReadCloser wraps ffmpeg stdout and calls cmd.Wait() on Close.
type previewReadCloser struct {
	io.ReadCloser
	cmd *exec.Cmd
}

func (p *previewReadCloser) Close() error {
	err := p.ReadCloser.Close()
	if p.cmd.Process != nil {
		_ = p.cmd.Process.Kill()
	}
	_ = p.cmd.Wait()
	return err
}
