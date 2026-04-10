package core

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

// ChannelJob tracks the state of a running or completed channel fetch.
type ChannelJob struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"` // "running" | "done" | "cancelled" | "error"
	Count     int       `json:"count"`
	Error     string    `json:"error,omitempty"`
	StartedAt time.Time `json:"startedAt"`
	cancel    context.CancelFunc
}

// ChannelJobRegistry manages in-memory channel fetch jobs.
type ChannelJobRegistry struct {
	mu   sync.RWMutex
	jobs map[string]*ChannelJob
}

// NewChannelJobRegistry creates a new registry.
func NewChannelJobRegistry() *ChannelJobRegistry {
	return &ChannelJobRegistry{jobs: make(map[string]*ChannelJob)}
}

// StartJob starts a new channel fetch job. onItem is called for each fetched video,
// onDone is called when the job completes. Returns the job ID.
func (r *ChannelJobRegistry) StartJob(url string, opts ChannelOpts, onItem func(VideoInfoLite), onDone func(total, errs int)) string {
	r.prune()

	id := newJobID()
	ctx, cancel := context.WithCancel(context.Background())
	job := &ChannelJob{
		ID:        id,
		Status:    "running",
		StartedAt: time.Now(),
		cancel:    cancel,
	}

	r.mu.Lock()
	r.jobs[id] = job
	r.mu.Unlock()

	go func() {
		items, errc := FetchChannelUploads(ctx, url, opts)
		errCount := 0
		count := 0
		for v := range items {
			onItem(v)
			count++
			r.mu.Lock()
			job.Count = count
			r.mu.Unlock()
		}
		if err := <-errc; err != nil {
			errCount++
		}

		r.mu.Lock()
		if ctx.Err() != nil {
			job.Status = "cancelled"
		} else {
			job.Status = "done"
		}
		r.mu.Unlock()

		cancel()
		onDone(count, errCount)
	}()

	return id
}

// CancelJob cancels a running job. Returns false if the job is not found.
func (r *ChannelJobRegistry) CancelJob(id string) bool {
	r.mu.RLock()
	job, ok := r.jobs[id]
	r.mu.RUnlock()
	if !ok {
		return false
	}
	job.cancel()
	return true
}

// GetJobStatus returns the current state of a job.
func (r *ChannelJobRegistry) GetJobStatus(id string) (*ChannelJob, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	job, ok := r.jobs[id]
	if !ok {
		return nil, false
	}
	cp := *job
	return &cp, true
}

// prune removes jobs older than 30 minutes.
func (r *ChannelJobRegistry) prune() {
	r.mu.Lock()
	defer r.mu.Unlock()
	cutoff := time.Now().Add(-30 * time.Minute)
	for id, job := range r.jobs {
		if job.StartedAt.Before(cutoff) {
			delete(r.jobs, id)
		}
	}
}

func newJobID() string {
	b := make([]byte, 6)
	rand.Read(b)
	return hex.EncodeToString(b)
}
