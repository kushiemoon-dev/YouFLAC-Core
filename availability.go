package core

import (
	"strings"
)

// AvailabilityResult describes the outcome of a YouTube availability check.
type AvailabilityResult struct {
	Available bool   `json:"available"`
	Reason    string `json:"reason,omitempty"` // removed, age_restricted, geo_blocked, private, unknown
	Title     string `json:"title,omitempty"`
	VideoID   string `json:"videoId,omitempty"`
}

// CheckAvailable validates a YouTube URL and probes metadata to detect
// age-restricted, geo-blocked, private, or removed videos.
func CheckAvailable(rawURL string) (AvailabilityResult, error) {
	videoID, err := ParseYouTubeURL(rawURL)
	if err != nil {
		return AvailabilityResult{Available: false, Reason: "invalid_url"}, err
	}

	info, err := GetVideoMetadata(videoID)
	if err != nil {
		return AvailabilityResult{
			Available: false,
			Reason:    classifyAvailabilityError(err),
			VideoID:   videoID,
		}, nil
	}
	return AvailabilityResult{
		Available: true,
		Title:     info.Title,
		VideoID:   videoID,
	}, nil
}

func classifyAvailabilityError(err error) string {
	if err == nil {
		return ""
	}
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "private"):
		return "private"
	case strings.Contains(msg, "age") && strings.Contains(msg, "confirm"):
		return "age_restricted"
	case strings.Contains(msg, "country") || strings.Contains(msg, "geo"):
		return "geo_blocked"
	case strings.Contains(msg, "removed") || strings.Contains(msg, "unavailable"):
		return "removed"
	default:
		return "unknown"
	}
}
