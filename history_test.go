package core

import (
	"testing"
)

func TestAddFromQueueItem_CarriesExplicit(t *testing.T) {
	SetDataDir(t.TempDir())

	h := NewHistory()
	item := &QueueItem{
		ID:       "test-1",
		Title:    "Song [Explicit]",
		Artist:   "Artist",
		Explicit: true,
		Status:   StatusComplete,
	}

	if err := h.AddFromQueueItem(item, "complete", ""); err != nil {
		t.Fatalf("AddFromQueueItem failed: %v", err)
	}

	entries := h.GetAll()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if !entries[0].Explicit {
		t.Errorf("expected Explicit=true, got false")
	}
}
