package core

import (
	"fmt"
	"regexp"
)

// MKV metadata embedding helpers

// GetMKVMetadataArgs returns ffmpeg args for embedding metadata
func GetMKVMetadataArgs(metadata *Metadata) []string {
	args := []string{}

	if metadata.Title != "" {
		args = append(args, "-metadata", fmt.Sprintf("title=%s", metadata.Title))
	}
	if metadata.Artist != "" {
		args = append(args, "-metadata", fmt.Sprintf("artist=%s", metadata.Artist))
	}
	if metadata.AlbumArtist != "" {
		args = append(args, "-metadata", fmt.Sprintf("album_artist=%s", metadata.AlbumArtist))
	}
	if metadata.Album != "" {
		args = append(args, "-metadata", fmt.Sprintf("album=%s", metadata.Album))
	}
	if metadata.Year > 0 {
		args = append(args, "-metadata", fmt.Sprintf("date=%d", metadata.Year))
	}
	if metadata.Genre != "" {
		args = append(args, "-metadata", fmt.Sprintf("genre=%s", metadata.Genre))
	}
	if metadata.ISRC != "" {
		args = append(args, "-metadata", fmt.Sprintf("isrc=%s", metadata.ISRC))
	}
	if metadata.Copyright != "" {
		args = append(args, "-metadata", fmt.Sprintf("copyright=%s", metadata.Copyright))
	}
	if metadata.Label != "" {
		args = append(args, "-metadata", fmt.Sprintf("publisher=%s", metadata.Label))
	}
	if metadata.DiscNumber > 0 {
		args = append(args, "-metadata", fmt.Sprintf("disc=%d", metadata.DiscNumber))
	}
	if metadata.TotalDiscs > 0 {
		args = append(args, "-metadata", fmt.Sprintf("totaldiscs=%d", metadata.TotalDiscs))
	}
	if metadata.YouTubeURL != "" {
		args = append(args, "-metadata", fmt.Sprintf("YOUTUBE_URL=%s", metadata.YouTubeURL))
	}
	if metadata.ViewCount > 0 {
		args = append(args, "-metadata", fmt.Sprintf("VIEW_COUNT=%d", metadata.ViewCount))
	}
	if len(metadata.UploadDate) == 8 {
		iso := metadata.UploadDate[:4] + "-" + metadata.UploadDate[4:6] + "-" + metadata.UploadDate[6:8]
		args = append(args, "-metadata", fmt.Sprintf("date=%s", iso))
	}

	return args
}

var explicitRegex = regexp.MustCompile(`(?i)(\[explicit\]|\(explicit\)|\bexplicit\s+version\b)`)

// DetectExplicit returns true when the title contains an explicit marker.
// Used for YouTube fallback where no explicit flag exists on the source.
func DetectExplicit(title string) bool {
	if title == "" {
		return false
	}
	return explicitRegex.MatchString(title)
}

