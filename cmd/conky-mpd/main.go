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
	Tracks []track

	CurrentIdx     int64
	CurrentElapsed float64
}

type track struct {
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

func format(data *mpdState) string {
	if !hasData(data) {
		return ""
	}
	r := "\nM P D\n${hr}\n"
	r += formatCurrentSong(data)
	r += formatUpcomingSongs(data)
	// addArt(&r, data)
	return r
}

func formatCurrentSong(allData *mpdState) string {
	data := allData.Tracks[int(allData.CurrentIdx)]
	r := ""
	addTrackInfo(&r, "Title", data.Title)
	addTrackInfo(&r, "Artists", data.Artist)
	addTrackInfo(&r, "Album", data.Album)
	// addTrackInfo(&r, "Genres", data.Genre)
	addTrackInfo(&r, "Playtime", fmt.Sprintf("%s / %s", asTimecode(allData.CurrentElapsed), asTimecode(data.Length)))
	addBar(&r, allData.CurrentElapsed/data.Length)
	return r
}

func formatUpcomingSongs(allData *mpdState) string {
	songsRemaining := len(allData.Tracks) - (int(allData.CurrentIdx) + 1)
	r := fmt.Sprintf("${hr}\n%d Songs left, next up...\n", songsRemaining)
	startIdx := int(allData.CurrentIdx) + 1
	// endIdx := startIdx + 5
	// if endIdx > len(allData.Tracks) {
	// 	endIdx = len(allData.Tracks)
	// }
	upcomingTracks := allData.Tracks[startIdx:]
	for _, track := range upcomingTracks {
		r += "•" + formatSongMinimal(&track)
	}
	return r
}

func formatSongMinimal(track *track) string {
	return fmt.Sprintf("%s | %s", track.Title, track.Artist) + "\n"
}

func hasData(data *mpdState) bool {
	return len(data.Tracks) > 0
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

func addBar(to *string, percent float64) {
	if percent < 0.0 {
		percent = 0.0
	} else if percent > 1.0 {
		percent = 1.0
	}
	val := percent * 100
	fmt.Fprintln(os.Stderr, val)
	*to += fmt.Sprintf("${execbar expr %.2f}\n", val)
}

func getData() (*mpdState, error) {
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
	fmt.Fprint(os.Stderr, status)

	result := mpdState{}

	for _, song := range allSongs {
		track := getTrack(song)
		result.Tracks = append(result.Tracks, *track)
	}

	result.CurrentIdx = strToInt64(status["song"])
	result.CurrentElapsed = strToFloat64(status["elapsed"])

	return &result, nil
}

func getTrack(metadata mpd.Attrs) *track {
	return &track{
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
	if int64(len(val)) <= l {
		return val
	}
	diff := int64(len(val)) - l
	unix := time.Now().Unix()
	step := unix % diff
	return val[step : l+step]
}

func etcetera(val string, l int64) string {
	if int64(len(val)) < l {
		return val
	}
	return val[:l-3] + "..."
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
