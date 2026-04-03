package core

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// Core is the central coordinator that ties all services together.
// It is created once at startup (via FFI) and provides the single
// entry-point for the RPC dispatcher.
type Core struct {
	config     *Config
	queue      *Queue
	history    *History
	downloader *UnifiedAudioDownloader
	lucida     *LucidaService
	tidalHifi  *TidalHifiService
	orpheus    *OrpheusDLService

	mu            sync.Mutex
	eventCallback func(event Event)
}

// NewCore initialises every subsystem and returns a ready-to-use Core.
// dataDir is the root directory for config, queue state, history, etc.
func NewCore(dataDir string) (*Core, error) {
	SetDataDir(dataDir)

	// Load config (missing file → defaults)
	config, err := LoadConfigWithEnv()
	if err != nil {
		slog.Warn("config load failed, using defaults", "err", err)
		config = GetDefaultConfig()
	}

	// HTTP timeout for downloads
	timeout := time.Duration(config.DownloadTimeoutMinutes) * time.Minute
	if timeout == 0 {
		timeout = 10 * time.Minute
	}

	// Fast HTTP client for metadata lookups (10s timeout, not 10 minutes)
	lookupClient, err := NewHTTPClient(10*time.Second, config.ProxyURL)
	if err != nil {
		slog.Warn("proxy config invalid, falling back to direct", "err", err)
		lookupClient, _ = NewHTTPClient(10*time.Second, "")
	}

	// Audio services (use fast lookup client)
	lucida := NewLucidaService(lookupClient)
	tidalHifi := NewTidalHifiService(lookupClient, config.PreferredQuality)
	orpheus := NewOrpheusDLService()

	// Unified downloader
	dlConfig := &DownloadConfig{
		PreferredFormat:  "flac",
		PreferredQuality: config.PreferredQuality,
		PlatformPriority: config.AudioSourcePriority,
		OutputDir:        config.OutputDirectory,
		Timeout:          timeout,
	}
	downloader := NewUnifiedAudioDownloader(dlConfig)

	// History
	history := NewHistory()

	// Queue
	concurrency := config.ConcurrentDownloads
	if concurrency < 1 {
		concurrency = 2
	}
	queue := NewQueue(context.Background(), concurrency)
	queue.SetConfig(config)
	queue.SetHistory(history)
	_ = queue.LoadQueue()
	queue.AutoSave(30 * time.Second)

	c := &Core{
		config:     config,
		queue:      queue,
		history:    history,
		downloader: downloader,
		lucida:     lucida,
		tidalHifi:  tidalHifi,
		orpheus:    orpheus,
	}

	// Wire queue progress events → Core eventCallback
	queue.SetProgressCallback(func(evt QueueEvent) {
		c.emitEvent(EventQueueChanged, evt)
	})

	return c, nil
}

// SetEventCallback registers the function that receives async events.
// Typically set once from the FFI layer.
func (c *Core) SetEventCallback(cb func(Event)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.eventCallback = cb
}

// emitEvent sends an event to the registered callback (if any).
func (c *Core) emitEvent(eventType string, payload interface{}) {
	c.mu.Lock()
	cb := c.eventCallback
	c.mu.Unlock()

	if cb != nil {
		cb(Event{Type: eventType, Payload: payload})
	}
}

// Shutdown persists state and stops background workers.
func (c *Core) Shutdown() {
	c.queue.StopProcessing()
	if err := c.queue.SaveQueue(); err != nil {
		slog.Error("failed to save queue on shutdown", "err", err)
	}
}
