package core

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestProcessItem_SetsSkippedStatus verifies that when the file index already
// contains a match at the exact target path, processItem marks the queue item
// with StatusSkipped (not StatusComplete) so the frontend can distinguish
// "newly downloaded" from "skipped because file already existed".
func TestProcessItem_SetsSkippedStatus(t *testing.T) {
	tmp := t.TempDir()

	// Use naming template "{artist} - {title}" so the target path is predictable.
	template := "{artist} - {title}"
	ext := ".mkv"

	meta := &Metadata{Title: "Title", Artist: "Artist"}
	targetPath := GenerateFilePath(meta, template, tmp, ext)

	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(targetPath, []byte("fake"), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	ctx := context.Background()
	q := NewQueue(ctx, 1)
	q.SetConfig(&Config{
		OutputDirectory: tmp,
		NamingTemplate:  template,
	})

	idx := NewFileIndex(t.TempDir())
	idx.AddEntry(FileIndexEntry{
		Path:      targetPath,
		Title:     "Title",
		Artist:    "Artist",
		IndexedAt: time.Now(),
	})
	q.SetFileIndex(idx)

	// AddToQueueWithMetadata presets Title+Artist so processItem skips the
	// YouTube fetch stage and goes straight to the file-index lookup.
	id, err := q.AddToQueueWithMetadata(
		DownloadRequest{VideoURL: ""},
		&VideoInfo{Title: "Title", Artist: "Artist"},
	)
	if err != nil {
		t.Fatalf("AddToQueueWithMetadata: %v", err)
	}

	// Call processItem directly instead of starting the worker loop.
	q.processItem(id)

	item := q.GetItem(id)
	if item == nil {
		t.Fatal("item not found after processItem")
	}
	if item.Status != StatusSkipped {
		t.Fatalf("expected StatusSkipped, got %q (stage=%q)", item.Status, item.Stage)
	}
	if item.OutputPath != targetPath {
		t.Errorf("expected OutputPath %q, got %q", targetPath, item.OutputPath)
	}
	if item.Progress != 100 {
		t.Errorf("expected progress 100, got %d", item.Progress)
	}
}
