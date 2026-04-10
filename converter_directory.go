package core

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// ConvertDirOptions holds parameters for a directory conversion job.
type ConvertDirOptions struct {
	Dir          string `json:"dir"`
	TargetFormat string `json:"targetFormat"`
	Bitrate      int    `json:"bitrate,omitempty"`
	SampleRate   int    `json:"sampleRate,omitempty"`
}

// DirConvertResult is emitted by ConvertDirectory for each file and as a final summary.
type DirConvertResult struct {
	SourcePath string `json:"sourcePath"`
	OutputPath string `json:"outputPath,omitempty"`
	Error      string `json:"error,omitempty"`
	Done       bool   `json:"done"`
	Total      int    `json:"total,omitempty"`
	Succeeded  int    `json:"succeeded,omitempty"`
	Failed     int    `json:"failed,omitempty"`
}

var audioExtensions = map[string]bool{
	".flac": true,
	".mp3":  true,
	".wav":  true,
	".m4a":  true,
	".aac":  true,
	".ogg":  true,
	".alac": true,
	".opus": true,
}

// ConvertDirectory walks dir, converts every audio file to opts.TargetFormat, and
// calls onResult for each file result and once at the end with a summary.
// Returns ctx.Err() if the context is cancelled, nil on successful completion,
// or an error if dir is invalid.
func ConvertDirectory(ctx context.Context, opts ConvertDirOptions, onResult func(DirConvertResult)) error {
	info, err := os.Stat(opts.Dir)
	if err != nil {
		return fmt.Errorf("directory not found: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", opts.Dir)
	}

	// Collect audio files
	var files []string
	if err := filepath.WalkDir(opts.Dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable entries
		}
		if d.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if audioExtensions[ext] {
			files = append(files, path)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("walk error: %w", err)
	}

	total := len(files)
	succeeded := 0
	failed := 0

	for _, srcPath := range files {
		// Check cancellation before each file
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		result, convErr := ConvertAudio(ConvertRequest{
			SourcePath:   srcPath,
			TargetFormat: opts.TargetFormat,
			Bitrate:      opts.Bitrate,
			SampleRate:   opts.SampleRate,
		})

		if convErr != nil {
			failed++
			onResult(DirConvertResult{
				SourcePath: srcPath,
				Error:      convErr.Error(),
			})
		} else {
			succeeded++
			onResult(DirConvertResult{
				SourcePath: srcPath,
				OutputPath: result.OutputPath,
			})
		}
	}

	onResult(DirConvertResult{
		Done:      true,
		Total:     total,
		Succeeded: succeeded,
		Failed:    failed,
	})

	return nil
}
