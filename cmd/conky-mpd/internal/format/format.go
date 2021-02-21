package format

import (
	"fmt"
	"time"

	"github.com/liampulles/bunsen-tools/cmd/conky-mpd/internal/domain"
)

// Format formats an MpdState into a string
func Format(data *domain.MpdState) string {
	if !hasData(data) {
		return ""
	}
	r := "\nM P D\n${hr}\n"
	r += formatCurrentSong(data)
	r += formatUpcomingSongs(data)
	// addArt(&r, data)
	return r
}

func formatCurrentSong(allData *domain.MpdState) string {
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

func formatUpcomingSongs(allData *domain.MpdState) string {
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

func formatSongMinimal(track *domain.Track) string {
	return fmt.Sprintf("%s | %s", track.Title, track.Artist) + "\n"
}

func hasData(data *domain.MpdState) bool {
	return len(data.Tracks) > 0
}

// func addArt(to *string, data *domain.MpdState) {
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
	*to += fmt.Sprintf("${execbar expr %.2f}\n", val)
}

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
