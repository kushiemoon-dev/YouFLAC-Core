package core

// Event represents an async event sent from Go core to the frontend (Flutter).
type Event struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload,omitempty"`
}

const (
	EventDownloadProgress = "download-progress"
	EventDownloadComplete = "download-complete"
	EventDownloadError    = "download-error"
	EventQueueChanged     = "queue-changed"
	EventServiceStatus    = "service-status-changed"
	EventLogMessage       = "log-message"
)
