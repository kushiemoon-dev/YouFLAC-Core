package core

import (
	"testing"
	"time"
)

func TestFetchCache_HitAndMiss(t *testing.T) {
	c := NewFetchCache(3, 100*time.Millisecond)

	if _, ok := c.Get("a"); ok {
		t.Fatal("expected miss")
	}
	c.Put("a", &VideoInfo{Title: "A"})
	v, ok := c.Get("a")
	if !ok {
		t.Fatal("expected hit")
	}
	if v.Title != "A" {
		t.Errorf("Title = %q", v.Title)
	}
}

func TestFetchCache_Expires(t *testing.T) {
	c := NewFetchCache(3, 10*time.Millisecond)
	c.Put("a", &VideoInfo{Title: "A"})
	time.Sleep(25 * time.Millisecond)
	if _, ok := c.Get("a"); ok {
		t.Fatal("expected expired miss")
	}
}

func TestFetchCache_LRUEviction(t *testing.T) {
	c := NewFetchCache(2, time.Hour)
	c.Put("a", &VideoInfo{Title: "A"})
	c.Put("b", &VideoInfo{Title: "B"})
	c.Get("a") // a becomes most-recent
	c.Put("c", &VideoInfo{Title: "C"})

	if _, ok := c.Get("b"); ok {
		t.Error("expected b to be evicted")
	}
	if _, ok := c.Get("a"); !ok {
		t.Error("expected a to survive")
	}
	if _, ok := c.Get("c"); !ok {
		t.Error("expected c to be present")
	}
}

func TestGetVideoMetadata_UsesCache(t *testing.T) {
	ConfigureFetchCache(true, 3600)
	defer ConfigureFetchCache(true, 3600)

	defaultFetchCache.Put("cachedID", &VideoInfo{ID: "cachedID", Title: "Cached"})
	got, err := GetVideoMetadata("cachedID")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if got.Title != "Cached" {
		t.Errorf("Title = %q, want Cached", got.Title)
	}
}
