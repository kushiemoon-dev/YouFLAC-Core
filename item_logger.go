package core

import (
	"context"
	"log/slog"
	"strings"
	"sync"
	"time"
)

const maxItemLogEntries = 500

type itemCtxKey struct{}

// WithItemID returns a copy of ctx tagged with the given queue item ID.
// Log records emitted with this context will be captured by the item's ring buffer.
func WithItemID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, itemCtxKey{}, id)
}

// ItemIDFromContext returns the item ID stored in ctx, or "" if none.
func ItemIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(itemCtxKey{}).(string); ok {
		return v
	}
	return ""
}

type itemRing struct {
	mu      sync.RWMutex
	entries []LogEntry
	seq     int64
}

func (r *itemRing) add(level, msg, fields string, t time.Time) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.seq++
	if t.IsZero() {
		t = time.Now()
	}
	e := LogEntry{
		ID:      r.seq,
		Time:    t.Format("15:04:05"),
		Level:   strings.ToUpper(level),
		Message: msg,
		Fields:  fields,
	}
	if len(r.entries) >= maxItemLogEntries {
		copy(r.entries, r.entries[1:])
		r.entries[len(r.entries)-1] = e
		return
	}
	r.entries = append(r.entries, e)
}

func (r *itemRing) snapshot() []LogEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]LogEntry, len(r.entries))
	copy(out, r.entries)
	return out
}

var (
	itemLoggersMu sync.RWMutex
	itemLoggers   = map[string]*itemRing{}
)

// RegisterItemLogger creates a new ring buffer for the given item ID.
func RegisterItemLogger(id string) {
	itemLoggersMu.Lock()
	defer itemLoggersMu.Unlock()
	itemLoggers[id] = &itemRing{}
}

// UnregisterItemLogger removes the ring buffer for the given item ID.
func UnregisterItemLogger(id string) {
	itemLoggersMu.Lock()
	defer itemLoggersMu.Unlock()
	delete(itemLoggers, id)
}

// GetItemLogs returns a snapshot of the entries captured for the given item ID.
func GetItemLogs(id string) []LogEntry {
	itemLoggersMu.RLock()
	ring, ok := itemLoggers[id]
	itemLoggersMu.RUnlock()
	if !ok {
		return nil
	}
	return ring.snapshot()
}

// ItemLogHandler is a slog.Handler that captures records tagged with an item ID
// into a per-item ring buffer before delegating to the wrapped handler.
type ItemLogHandler struct {
	next slog.Handler
}

// NewItemLogHandler wraps next so that records carrying an item ID in their
// context are also recorded in the matching item ring buffer.
func NewItemLogHandler(next slog.Handler) *ItemLogHandler {
	return &ItemLogHandler{next: next}
}

func (h *ItemLogHandler) Enabled(ctx context.Context, l slog.Level) bool {
	// If an item is being tracked, always enable so per-item debug logs are captured,
	// regardless of the global level.
	if ItemIDFromContext(ctx) != "" {
		return true
	}
	return h.next.Enabled(ctx, l)
}

func (h *ItemLogHandler) Handle(ctx context.Context, r slog.Record) error {
	if id := ItemIDFromContext(ctx); id != "" {
		itemLoggersMu.RLock()
		ring, ok := itemLoggers[id]
		itemLoggersMu.RUnlock()
		if ok {
			ring.add(r.Level.String(), r.Message, formatAttrs(r), r.Time)
		}
		// Only forward to the next handler if it would have accepted this level.
		if !h.next.Enabled(ctx, r.Level) {
			return nil
		}
	}
	return h.next.Handle(ctx, r)
}

func (h *ItemLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &ItemLogHandler{next: h.next.WithAttrs(attrs)}
}

func (h *ItemLogHandler) WithGroup(name string) slog.Handler {
	return &ItemLogHandler{next: h.next.WithGroup(name)}
}

// formatAttrs flattens a slog.Record's attributes into a space-separated
// "key=value" string for inclusion in the per-item ring buffer.
func formatAttrs(r slog.Record) string {
	if r.NumAttrs() == 0 {
		return ""
	}
	var b strings.Builder
	first := true
	r.Attrs(func(a slog.Attr) bool {
		if !first {
			b.WriteByte(' ')
		}
		first = false
		b.WriteString(a.Key)
		b.WriteByte('=')
		b.WriteString(a.Value.String())
		return true
	})
	return b.String()
}
