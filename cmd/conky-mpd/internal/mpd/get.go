package mpd

import (
	"fmt"
	"strconv"

	"github.com/fhs/gompd/v2/mpd"
	"github.com/liampulles/bunsen-tools/cmd/conky-mpd/internal/domain"
)

// GetMpdState polls MPD for current information
func GetMpdState() (*domain.MpdState, error) {
	conn, err := mpd.Dial("tcp", "127.0.0.1:6600")
	if err != nil {
		return nil, fmt.Errorf("session connect error: %w", err)
	}
	defer conn.Close()

	allSongs, err := conn.PlaylistInfo(-1, -1)
	if err != nil {
		return nil, fmt.Errorf("mpd currentSong error: %w", err)
	}
	status, err := conn.Status()
	if err != nil {
		return nil, fmt.Errorf("mpd status error: %w", err)
	}

	result := domain.MpdState{}

	for _, song := range allSongs {
		track := getTrack(song)
		result.Tracks = append(result.Tracks, *track)
	}

	result.CurrentIdx = strToInt64(status["song"])
	result.CurrentElapsed = strToFloat64(status["elapsed"])

	return &result, nil
}

func getTrack(metadata mpd.Attrs) *domain.Track {
	return &domain.Track{
		Title:       metadata["Title"],
		Album:       metadata["Album"],
		Artist:      metadata["Artist"],
		Genre:       metadata["Genre"],
		TrackNumber: strToInt64(metadata["Track"]),
		DiscNumber:  strToInt64(metadata["DiscNumber"]),
		Year:        metadata["OriginalDate"],

		Length: strToFloat64(metadata["duration"]),

		// ArtURL: safeAsString(metadata["mpris:artUrl"]),
	}
}

func strToInt64(val string) int64 {
	i64, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0
	}
	return i64
}

func strToFloat64(val string) float64 {
	f64, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return 0.0
	}
	return f64
}
