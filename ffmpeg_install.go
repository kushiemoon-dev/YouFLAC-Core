package core

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// FFmpegInfo describes the currently available ffmpeg binary.
type FFmpegInfo struct {
	Path    string `json:"path"`
	Version string `json:"version"`
	Found   bool   `json:"found"`
	Source  string `json:"source"` // path, bundled, common, missing
}

// FFmpegProgress is emitted during InstallFFmpeg.
type FFmpegProgress struct {
	Downloaded int64  `json:"downloaded"`
	Total      int64  `json:"total"`
	Stage      string `json:"stage"` // download, extract, install, done, error
	Error      string `json:"error,omitempty"`
}

// FFmpegProgressCallback receives installation progress events.
type FFmpegProgressCallback func(p FFmpegProgress)

// DetectFFmpeg searches PATH, bundled bin dir, and common OS locations.
func DetectFFmpeg() FFmpegInfo {
	type candidate struct {
		path   string
		source string
	}
	candidates := []candidate{
		{filepath.Join(GetBinPath(), "ffmpeg"), "bundled"},
		{filepath.Join(GetBinPath(), "ffmpeg.exe"), "bundled"},
	}
	switch runtime.GOOS {
	case "linux", "darwin":
		candidates = append(candidates,
			candidate{"/usr/bin/ffmpeg", "common"},
			candidate{"/usr/local/bin/ffmpeg", "common"},
		)
		if home, err := os.UserHomeDir(); err == nil {
			candidates = append(candidates, candidate{filepath.Join(home, "bin", "ffmpeg"), "common"})
		}
	case "windows":
		candidates = append(candidates, candidate{`C:\ffmpeg\bin\ffmpeg.exe`, "common"})
	}
	for _, c := range candidates {
		if fileExists(c.path) {
			return FFmpegInfo{Path: c.path, Version: readFFmpegVersion(c.path), Found: true, Source: c.source}
		}
	}
	if p, err := exec.LookPath("ffmpeg"); err == nil {
		return FFmpegInfo{Path: p, Version: readFFmpegVersion(p), Found: true, Source: "path"}
	}
	return FFmpegInfo{Path: "ffmpeg", Found: false, Source: "missing"}
}

func readFFmpegVersion(path string) string {
	cmd := exec.Command(path, "-version")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	line := string(out)
	if idx := strings.IndexByte(line, '\n'); idx > 0 {
		line = line[:idx]
	}
	return line
}

// ffmpegDownloadURL returns the static build URL for the current platform.
func ffmpegDownloadURL() (string, error) {
	switch runtime.GOOS + "/" + runtime.GOARCH {
	case "linux/amd64":
		return "https://johnvansickle.com/ffmpeg/releases/ffmpeg-release-amd64-static.tar.xz", nil
	case "linux/arm64":
		return "https://johnvansickle.com/ffmpeg/releases/ffmpeg-release-arm64-static.tar.xz", nil
	case "darwin/amd64", "darwin/arm64":
		return "https://evermeet.cx/ffmpeg/getrelease/ffmpeg/zip", nil
	case "windows/amd64":
		return "https://www.gyan.dev/ffmpeg/builds/ffmpeg-release-essentials.zip", nil
	default:
		return "", fmt.Errorf("unsupported platform: %s/%s", runtime.GOOS, runtime.GOARCH)
	}
}

// InstallFFmpeg downloads a static ffmpeg binary archive under GetBinPath().
// onProgress is called with cumulative byte counters during the download stage.
// Extraction is NOT yet implemented — this phase ships download-with-progress only.
func InstallFFmpeg(ctx context.Context, onProgress FFmpegProgressCallback) error {
	if onProgress == nil {
		onProgress = func(FFmpegProgress) {}
	}
	url, err := ffmpegDownloadURL()
	if err != nil {
		return err
	}
	binDir := GetBinPath()
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		return err
	}
	archive := filepath.Join(binDir, "ffmpeg-download.bin")
	onProgress(FFmpegProgress{Stage: "download"})
	if err := downloadWithProgress(ctx, url, archive, func(done, total int64) {
		onProgress(FFmpegProgress{Stage: "download", Downloaded: done, Total: total})
	}); err != nil {
		return err
	}
	onProgress(FFmpegProgress{Stage: "done"})
	return nil
}

// downloadWithProgress streams url to dest and invokes cb after every chunk.
func downloadWithProgress(ctx context.Context, url, dest string, cb func(done, total int64)) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("download failed: %s", resp.Status)
	}
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	total := resp.ContentLength
	var done int64
	buf := make([]byte, 32*1024)
	for {
		n, rerr := resp.Body.Read(buf)
		if n > 0 {
			if _, werr := f.Write(buf[:n]); werr != nil {
				return werr
			}
			done += int64(n)
			cb(done, total)
		}
		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			return rerr
		}
	}
	return nil
}

var _ = io.Discard
