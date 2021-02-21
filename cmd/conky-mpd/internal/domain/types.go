package domain

// MpdState describes the current state of MPD
type MpdState struct {
	Tracks []Track

	CurrentIdx     int64
	CurrentElapsed float64
}

// Track describes a music track's metadata
type Track struct {
	Title       string
	Album       string
	Artist      string
	Genre       string
	TrackNumber int64
	DiscNumber  int64
	Year        string
	Length      float64
	ArtURL      string
}
