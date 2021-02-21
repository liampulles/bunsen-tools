package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/fhs/gompd/mpd"
)

func main() {
	data, err := getData()
	if err != nil {
		// Quit silently
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err.Error())
		os.Exit(1)
	}
	formatted := format(data)
	fmt.Print(formatted)
	os.Exit(0)
}

type mpdState struct {
	CurrentTrack currentTrack
}

type currentTrack struct {
	Title       string
	Album       string
	Artist      string
	Genre       string
	TrackNumber int64
	DiscNumber  int64
	Year        string

	Position float64
	Length   float64

	ArtURL string
}

func format(data *mpdState) string {
	if !hasData(data) {
		return ""
	}
	r := "\nM P D\n${hr}\n"
	// addArt(&r, data)
	addTrackInfo(&r, "Title", data.CurrentTrack.Title)
	addTrackInfo(&r, "Artists", data.CurrentTrack.Artist)
	addTrackInfo(&r, "Album", data.CurrentTrack.Album)
	addTrackInfo(&r, "Genres", data.CurrentTrack.Genre)
	addTrackInfo(&r, "Playtime", fmt.Sprintf("%s / %s", asTimecode(data.CurrentTrack.Position), asTimecode(data.CurrentTrack.Length)))
	return r
}

func hasData(data *mpdState) bool {
	return data.CurrentTrack.Title != ""
}

// func addArt(to *string, data *mpdState) {
// 	if data.CurrentTrack.ArtURL == "" {
// 		return
// 	}
// 	*to += fmt.Sprintf("${image %s -p 0,135 -s 200x200}${voffset 200}\n",
// 		asPath(data.CurrentTrack.ArtURL))
// }

func addTrackInfo(to *string, name string, data string) {

	if data == "" {
		return
	}
	*to += fmt.Sprintf("%s:${alignr}%s\n",
		name, scroll(data, 20))
}

func getData() (*mpdState, error) {
	conn, err := mpd.Dial("tcp", "127.0.0.1:6600")
	if err != nil {
		return nil, fmt.Errorf("session connect error: %w", err)
	}
	defer conn.Close()

	// getTracks(conn)

	currentTrack, err := getCurrentTrack(conn)
	if err != nil {
		return nil, fmt.Errorf("could not get current track: %w", err)
	}

	return &mpdState{
		CurrentTrack: *currentTrack,
	}, nil
}

func getCurrentTrack(conn *mpd.Client) (*currentTrack, error) {
	metadata, err := conn.CurrentSong()
	if err != nil {
		return nil, fmt.Errorf("mpd currentSong error: %w", err)
	}
	status, err := conn.Status()
	if err != nil {
		return nil, fmt.Errorf("mpd status error: %w", err)
	}

	// allSongs, _ := conn.PlaylistInfo(-1, -1)
	// fmt.Fprintf(os.Stderr, "%+v\n", metadata["elapsed"])
	fmt.Fprintf(os.Stderr, "%+v\n", metadata)

	return &currentTrack{
		Title:       metadata["Title"],
		Album:       metadata["Album"],
		Artist:      metadata["Artist"],
		Genre:       metadata["Genre"],
		TrackNumber: strToInt64(metadata["Track"]),
		DiscNumber:  strToInt64(metadata["DiscNumber"]),
		Year:        metadata["OriginalDate"],

		Length:   strToFloat64(metadata["duration"]),
		Position: strToFloat64(status["elapsed"]),

		// ArtURL: safeAsString(metadata["mpris:artUrl"]),
	}, nil
}

// func getTracks(conn *mpd.Client) {
// 	var val interface{}
// 	err := conn.Object("org.mpris.MediaPlayer2.mpd", "/org/mpris/MediaPlayer2").Call("org.freedesktop.DBus.Properties.Get", 0, "org.mpris.MediaPlayer2.TrackList", "Tracks").Store(&val)
// 	if err != nil {
// 		fmt.Fprintf(os.Stderr, "could not get tracks: %s\n", err.Error())
// 	}
// 	fmt.Fprintf(os.Stderr, "%+v", val)
// }

func asTimecode(val float64) string {
	intVal := int64(val)
	mins := intVal / 60
	secs := intVal % 60
	return fmt.Sprintf("%01d:%02d", mins, secs)
}

func scroll(val string, l int64) string {
	if int64(len(val)) < l {
		return val
	}
	repeated := val + "/"
	repeated = repeated + repeated
	repeated = repeated + repeated
	unix := time.Now().Unix()
	step := unix % (l + 1)
	return repeated[step : l+step]
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
