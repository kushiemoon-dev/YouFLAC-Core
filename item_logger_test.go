package core

import (
	"context"
	"log/slog"
	"strings"
	"testing"
)

func TestItemLogger_CapturesTaggedEntries(t *testing.T) {
	RegisterItemLogger("item-1")
	defer UnregisterItemLogger("item-1")

	h := NewItemLogHandler(slog.NewTextHandler(discardWriter{}, nil))
	logger := slog.New(h)

	ctx := WithItemID(context.Background(), "item-1")
	logger.InfoContext(ctx, "fetching info", "stage", "fetch")
	logger.DebugContext(ctx, "resolving url")

	entries := GetItemLogs("item-1")
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Message != "fetching info" {
		t.Errorf("entry[0].Message = %q", entries[0].Message)
	}
	if entries[0].Level != "INFO" {
		t.Errorf("entry[0].Level = %q", entries[0].Level)
	}
}

type discardWriter struct{}

func (discardWriter) Write(p []byte) (int, error) { return len(p), nil }

func TestItemLogger_RingBufferCap(t *testing.T) {
	RegisterItemLogger("cap")
	defer UnregisterItemLogger("cap")

	h := NewItemLogHandler(slog.NewTextHandler(discardWriter{}, nil))
	logger := slog.New(h)
	ctx := WithItemID(context.Background(), "cap")

	for i := 0; i < 600; i++ {
		logger.InfoContext(ctx, "msg")
	}
	entries := GetItemLogs("cap")
	if len(entries) != 500 {
		t.Fatalf("expected 500 entries, got %d", len(entries))
	}
	if entries[0].ID != 101 {
		t.Errorf("expected first ID 101 after rotation, got %d", entries[0].ID)
	}
}

func TestItemLogger_CapturesAttrs(t *testing.T) {
	RegisterItemLogger("attrs")
	defer UnregisterItemLogger("attrs")

	h := NewItemLogHandler(slog.NewTextHandler(discardWriter{}, nil))
	logger := slog.New(h)
	ctx := WithItemID(context.Background(), "attrs")

	logger.InfoContext(ctx, "fetching", "stage", "fetch", "source", "tidal")

	entries := GetItemLogs("attrs")
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if !strings.Contains(entries[0].Fields, "stage=fetch") {
		t.Errorf("expected fields to contain stage=fetch, got %q", entries[0].Fields)
	}
	if !strings.Contains(entries[0].Fields, "source=tidal") {
		t.Errorf("expected fields to contain source=tidal, got %q", entries[0].Fields)
	}
}

func TestItemLogger_DebugCapturedWhenBaseIsInfo(t *testing.T) {
	RegisterItemLogger("debug-cap")
	defer UnregisterItemLogger("debug-cap")

	// Base handler at INFO — DEBUG would normally be dropped.
	base := slog.NewTextHandler(discardWriter{}, &slog.HandlerOptions{Level: slog.LevelInfo})
	h := NewItemLogHandler(base)
	logger := slog.New(h)
	ctx := WithItemID(context.Background(), "debug-cap")

	logger.DebugContext(ctx, "low-level trace")
	logger.InfoContext(ctx, "normal event")

	entries := GetItemLogs("debug-cap")
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries (debug + info) captured in ring, got %d", len(entries))
	}
	if entries[0].Level != "DEBUG" {
		t.Errorf("expected first entry level DEBUG, got %q", entries[0].Level)
	}
}

func TestItemLogger_UntaggedRecordsSkipRing(t *testing.T) {
	RegisterItemLogger("tagged")
	defer UnregisterItemLogger("tagged")

	h := NewItemLogHandler(slog.NewTextHandler(discardWriter{}, nil))
	logger := slog.New(h)

	// No WithItemID — untagged context.
	logger.InfoContext(context.Background(), "global event")

	entries := GetItemLogs("tagged")
	if len(entries) != 0 {
		t.Errorf("expected 0 entries (no item id in ctx), got %d", len(entries))
	}
}
