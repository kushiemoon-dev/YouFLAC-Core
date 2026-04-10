package core

import (
	"context"
	"log/slog"
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
