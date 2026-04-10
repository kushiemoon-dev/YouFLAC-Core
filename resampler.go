package core

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ResampleOptions configures an audio resampling operation.
type ResampleOptions struct {
	InputPath  string `json:"inputPath"`
	OutputPath string `json:"outputPath"`
	SampleRate int    `json:"sampleRate"`
	BitDepth   int    `json:"bitDepth"`
	Dither     bool   `json:"dither"`
	Format     string `json:"format"`
}

// SupportedSampleRates lists the sample rates accepted by Resample.
var SupportedSampleRates = []int{44100, 48000, 88200, 96000, 176400, 192000}

// SupportedBitDepths lists the bit depths accepted by Resample.
var SupportedBitDepths = []int{16, 24, 32}

// SupportedResampleFormats lists the output formats accepted by Resample.
var SupportedResampleFormats = []string{"flac", "wav", "alac"}

func intInSlice(n int, s []int) bool {
	for _, v := range s {
		if v == n {
			return true
		}
	}
	return false
}

func strInSlice(v string, s []string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}

// Resample converts an audio file to the specified sample rate and bit depth.
func Resample(opts ResampleOptions) error {
	if opts.InputPath == "" {
		return fmt.Errorf("inputPath required")
	}
	if _, err := os.Stat(opts.InputPath); err != nil {
		return fmt.Errorf("input not found: %w", err)
	}
	if opts.OutputPath == "" {
		return fmt.Errorf("outputPath required")
	}
	if !intInSlice(opts.SampleRate, SupportedSampleRates) {
		return fmt.Errorf("unsupported sample rate: %d", opts.SampleRate)
	}
	if !intInSlice(opts.BitDepth, SupportedBitDepths) {
		return fmt.Errorf("unsupported bit depth: %d", opts.BitDepth)
	}
	format := strings.ToLower(opts.Format)
	if !strInSlice(format, SupportedResampleFormats) {
		return fmt.Errorf("unsupported format: %s", opts.Format)
	}
	if err := os.MkdirAll(filepath.Dir(opts.OutputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output dir: %w", err)
	}

	sampleFmt, codec := "s16", "pcm_s16le"
	switch opts.BitDepth {
	case 24:
		sampleFmt, codec = "s32", "pcm_s24le"
	case 32:
		sampleFmt, codec = "s32", "pcm_s32le"
	}

	af := fmt.Sprintf("aresample=resampler=soxr:precision=28:osr=%d", opts.SampleRate)
	if opts.Dither {
		af += ":dither_method=triangular"
	}

	args := []string{"-y", "-i", opts.InputPath, "-af", af, "-ar", fmt.Sprintf("%d", opts.SampleRate), "-sample_fmt", sampleFmt}
	switch format {
	case "flac":
		args = append(args, "-c:a", "flac")
	case "wav":
		args = append(args, "-c:a", codec)
	case "alac":
		args = append(args, "-c:a", "alac")
	}
	args = append(args, "-vn", opts.OutputPath)

	cmd := exec.Command(GetFFmpegPath(), args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg resample failed: %v - %s", err, stderr.String())
	}
	return nil
}
