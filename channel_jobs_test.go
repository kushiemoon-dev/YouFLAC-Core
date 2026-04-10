package core

import (
	"sync"
	"testing"
	"time"
)

func TestChannelJobs_StartAndStream(t *testing.T) {
	SetYtdlpBinaryForTests("./testdata/ytdlp_channel_ok.sh")
	t.Cleanup(func() { SetYtdlpBinaryForTests("yt-dlp") })

	reg := NewChannelJobRegistry()
	done := make(chan struct{})
	var count int

	id := reg.StartJob("https://www.youtube.com/@Test", ChannelOpts{}, func(_ string, _ VideoInfoLite, _ int) {
		count++
	}, func(total, errs int) {
		close(done)
	})

	if id == "" {
		t.Fatal("expected non-empty job ID")
	}

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("job did not complete in time")
	}

	if count != 10 {
		t.Errorf("expected 10 items, got %d", count)
	}

	job, ok := reg.GetJobStatus(id)
	if !ok {
		t.Fatal("job not found after completion")
	}
	if job.Status != "done" {
		t.Errorf("expected status=done, got %q", job.Status)
	}
}

func TestChannelJobs_CancelJob(t *testing.T) {
	SetYtdlpBinaryForTests("./testdata/ytdlp_channel_ok.sh")
	t.Cleanup(func() { SetYtdlpBinaryForTests("yt-dlp") })

	reg := NewChannelJobRegistry()
	started := make(chan struct{})
	var mu sync.Mutex
	var count int

	id := reg.StartJob("https://www.youtube.com/@Test", ChannelOpts{}, func(_ string, _ VideoInfoLite, _ int) {
		mu.Lock()
		count++
		if count == 1 {
			close(started)
		}
		mu.Unlock()
	}, func(_, _ int) {})

	<-started
	ok := reg.CancelJob(id)
	if !ok {
		t.Fatal("CancelJob returned false for running job")
	}

	// Give cancellation time to propagate
	time.Sleep(100 * time.Millisecond)

	job, found := reg.GetJobStatus(id)
	if !found {
		t.Fatal("job not found")
	}
	if job.Status != "cancelled" && job.Status != "done" {
		t.Errorf("expected cancelled or done, got %q", job.Status)
	}
}

func TestChannelJobs_UnknownIDNotFound(t *testing.T) {
	reg := NewChannelJobRegistry()
	_, ok := reg.GetJobStatus("nonexistent-id")
	if ok {
		t.Error("expected not found for unknown id")
	}
}

func TestChannelJobs_CancelUnknownID(t *testing.T) {
	reg := NewChannelJobRegistry()
	ok := reg.CancelJob("nonexistent")
	if ok {
		t.Error("expected false for unknown id")
	}
}

func TestChannelJobs_ConcurrentStartUniqueIDs(t *testing.T) {
	SetYtdlpBinaryForTests("./testdata/ytdlp_channel_ok.sh")
	t.Cleanup(func() { SetYtdlpBinaryForTests("yt-dlp") })

	reg := NewChannelJobRegistry()
	var mu sync.Mutex
	ids := make(map[string]bool)
	var wg sync.WaitGroup

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			id := reg.StartJob("https://www.youtube.com/@Test", ChannelOpts{MaxItems: 1}, func(_ string, _ VideoInfoLite, _ int) {}, func(_, _ int) {})
			mu.Lock()
			ids[id] = true
			mu.Unlock()
		}()
	}
	wg.Wait()

	if len(ids) != 5 {
		t.Errorf("expected 5 unique IDs, got %d", len(ids))
	}
}
