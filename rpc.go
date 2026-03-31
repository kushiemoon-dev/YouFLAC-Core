package core

import (
	"encoding/json"
	"fmt"
	"strings"
)

// RPCRequest is the JSON envelope received from the Flutter side.
type RPCRequest struct {
	Method string          `json:"method"`
	Params json.RawMessage `json:"params,omitempty"`
}

// RPCResponse is the JSON envelope sent back to Flutter.
type RPCResponse struct {
	Result interface{} `json:"result,omitempty"`
	Error  *RPCError   `json:"error,omitempty"`
}

// RPCError carries a machine-readable code and human-readable message.
type RPCError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// HandleRPC parses a JSON-RPC request and returns a JSON response string.
func (c *Core) HandleRPC(input string) string {
	var req RPCRequest
	if err := json.Unmarshal([]byte(input), &req); err != nil {
		return marshalError("parse_error", fmt.Sprintf("invalid JSON: %v", err))
	}

	result, err := c.dispatch(req.Method, req.Params)
	if err != nil {
		return marshalError("method_error", err.Error())
	}
	return marshalResult(result)
}

// dispatch routes a method name to the appropriate handler.
func (c *Core) dispatch(method string, params json.RawMessage) (interface{}, error) {
	switch method {

	// ── Config ───────────────────────────────────────────────────────────
	case "getConfig":
		return c.config, nil

	case "saveConfig":
		var cfg Config
		if err := json.Unmarshal(params, &cfg); err != nil {
			return nil, fmt.Errorf("invalid config params: %w", err)
		}
		if err := SaveConfig(&cfg); err != nil {
			return nil, err
		}
		c.config = &cfg
		c.queue.SetConfig(&cfg)
		return map[string]bool{"ok": true}, nil

	// ── Resolve / Download ──────────────────────────────────────────────
	case "fetchContent":
		var p struct {
			URL string `json:"url"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, fmt.Errorf("invalid params: %w", err)
		}
		return c.fetchContent(p.URL)

	case "resolveUrl":
		var p struct {
			URL string `json:"url"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, fmt.Errorf("invalid params: %w", err)
		}
		return c.downloader.GetTrackInfo(p.URL)

	case "download":
		var req DownloadRequest
		if err := json.Unmarshal(params, &req); err != nil {
			return nil, fmt.Errorf("invalid params: %w", err)
		}
		id, err := c.queue.AddToQueue(req)
		if err != nil {
			return nil, err
		}
		c.queue.StartProcessing()
		return map[string]string{"id": id}, nil

	// ── Download control ────────────────────────────────────────────────
	case "download.pause":
		id, err := extractID(params)
		if err != nil {
			return nil, err
		}
		return nil, c.queue.PauseItem(id)

	case "download.resume":
		id, err := extractID(params)
		if err != nil {
			return nil, err
		}
		return nil, c.queue.ResumeItem(id)

	case "download.cancel":
		id, err := extractID(params)
		if err != nil {
			return nil, err
		}
		return nil, c.queue.CancelItem(id)

	case "download.retry":
		var p struct {
			ID       string               `json:"id"`
			Override *RetryOverrideRequest `json:"override,omitempty"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, fmt.Errorf("invalid params: %w", err)
		}
		if p.Override != nil {
			return c.queue.RetryWithOverride(p.ID, *p.Override)
		}
		// Simple retry: reset to pending
		item := c.queue.GetItem(p.ID)
		if item == nil {
			return nil, fmt.Errorf("item not found: %s", p.ID)
		}
		return c.queue.RetryWithOverride(p.ID, RetryOverrideRequest{})

	// ── Queue ───────────────────────────────────────────────────────────
	case "queue.list":
		return c.queue.GetQueue(), nil

	case "queue.stats":
		return c.queue.GetStats(), nil

	case "queue.clear":
		c.queue.ClearAll()
		return map[string]bool{"ok": true}, nil

	case "queue.retryAllFailed":
		n := c.queue.RetryFailed()
		if n > 0 {
			c.queue.StartProcessing()
		}
		return map[string]int{"retried": n}, nil

	case "queue.exportFailed":
		failed := c.queue.GetFailedItems()
		var lines []string
		for _, item := range failed {
			line := item.VideoURL
			if item.Title != "" {
				line = fmt.Sprintf("%s | %s", item.Title, item.VideoURL)
			}
			if item.Error != "" {
				line += " | " + item.Error
			}
			lines = append(lines, line)
		}
		return map[string]string{"text": strings.Join(lines, "\n")}, nil

	case "queue.persist":
		return nil, c.queue.SaveQueue()

	// ── History ─────────────────────────────────────────────────────────
	case "history.list":
		return c.history.GetAll(), nil

	case "history.clear":
		return nil, c.history.Clear()

	// ── Services ────────────────────────────────────────────────────────
	case "services.status":
		return CheckServiceStatus(c.config.ProxyURL), nil

	// ── Converter ───────────────────────────────────────────────────────
	case "convert":
		var req ConvertRequest
		if err := json.Unmarshal(params, &req); err != nil {
			return nil, fmt.Errorf("invalid params: %w", err)
		}
		return ConvertAudio(req)

	case "convert.formats":
		return SupportedConvertFormats, nil

	// ── Playlist ────────────────────────────────────────────────────────
	case "playlist.generate":
		var p struct {
			OutputDir    string `json:"outputDir"`
			PlaylistName string `json:"playlistName"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, fmt.Errorf("invalid params: %w", err)
		}
		completed := c.queue.GetQueue()
		var items []QueueItem
		for _, it := range completed {
			if it.Status == StatusComplete {
				items = append(items, it)
			}
		}
		return nil, GenerateM3U8(items, p.OutputDir, p.PlaylistName)

	// ── Meta ────────────────────────────────────────────────────────────
	case "getVersion":
		return "1.0.0-mobile", nil

	default:
		return nil, fmt.Errorf("unknown method: %s", method)
	}
}

// fetchContentResponse is the envelope returned by the fetchContent RPC.
type fetchContentResponse struct {
	Type       string               `json:"type"`
	Title      string               `json:"title"`
	Creator    string               `json:"creator"`
	CoverURL   string               `json:"coverUrl,omitempty"`
	TrackCount int                  `json:"trackCount"`
	Tracks     []fetchContentTrack  `json:"tracks"`
}

type fetchContentTrack struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Artist      string  `json:"artist"`
	Duration    float64 `json:"duration"`
	TrackNumber int     `json:"trackNumber"`
}

// fetchContent resolves a music URL and returns track metadata.
// Resolution chain: Lucida → Odesli+TidalHiFi → direct service fallback.
func (c *Core) fetchContent(rawURL string) (*fetchContentResponse, error) {
	if rawURL == "" {
		return nil, fmt.Errorf("url is required")
	}

	// 1) Try Lucida first (supports multi-platform URLs natively).
	if c.lucida.IsAvailable() {
		if info, err := c.lucida.GetTrackInfo(rawURL); err == nil && isValidTrackInfo(info) {
			return trackInfoToResponse(info), nil
		}
	}

	// 2) Try Odesli (song.link) → resolve to best FLAC source → TidalHiFi.
	if resolved, err := ResolveMusicURL(rawURL); err == nil {
		if _, tidalURL := GetBestFLACSource(resolved); tidalURL != "" {
			if c.tidalHifi.IsAvailable() {
				if info, err := c.tidalHifi.GetTrackInfo(tidalURL); err == nil && isValidTrackInfo(info) {
					return trackInfoToResponse(info), nil
				}
			}
		}
	}

	// 3) Fallback: try each service directly.
	services := []AudioDownloadService{c.lucida, c.tidalHifi, c.orpheus}
	for _, svc := range services {
		if svc == nil || !svc.IsAvailable() {
			continue
		}
		if info, err := svc.GetTrackInfo(rawURL); err == nil && isValidTrackInfo(info) {
			return trackInfoToResponse(info), nil
		}
	}

	return nil, fmt.Errorf("no service could resolve content for URL: %s", rawURL)
}

// trackInfoToResponse converts an AudioTrackInfo into the fetchContent envelope.
func trackInfoToResponse(info *AudioTrackInfo) *fetchContentResponse {
	coverURL := info.CoverURL
	// Validate cover URL — reject malformed ones (e.g. empty album cover from Tidal)
	if strings.Contains(coverURL, "//640x640") || strings.Contains(coverURL, "/images//") {
		coverURL = ""
	}
	return &fetchContentResponse{
		Type:       "track",
		Title:      info.Title,
		Creator:    info.Artist,
		CoverURL:   coverURL,
		TrackCount: 1,
		Tracks: []fetchContentTrack{
			{
				ID:          info.ID,
				Title:       info.Title,
				Artist:      info.Artist,
				Duration:    info.Duration,
				TrackNumber: info.TrackNumber,
			},
		},
	}
}

// isValidTrackInfo checks that the resolved track has meaningful data.
func isValidTrackInfo(info *AudioTrackInfo) bool {
	return info != nil && info.Title != ""
}

// extractID is a small helper to pull an "id" field from params.
func extractID(params json.RawMessage) (string, error) {
	var p struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return "", fmt.Errorf("invalid params: %w", err)
	}
	if p.ID == "" {
		return "", fmt.Errorf("missing id")
	}
	return p.ID, nil
}

func marshalResult(result interface{}) string {
	resp := RPCResponse{Result: result}
	data, _ := json.Marshal(resp)
	return string(data)
}

func marshalError(code, message string) string {
	resp := RPCResponse{Error: &RPCError{Code: code, Message: message}}
	data, _ := json.Marshal(resp)
	return string(data)
}
