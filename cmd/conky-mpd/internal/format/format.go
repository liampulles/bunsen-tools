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
	r := fmt.Sprintf("\n%sM P D ${hr}\n", c(0))
	r += formatCurrentSong(data)
	r += formatUpcomingSongs(data)
	// addArt(&r, data)
	return r
}

func formatCurrentSong(allData *domain.MpdState) string {
	data := allData.Tracks[int(allData.CurrentIdx)]
	r := ""
	addArt(&r)
	addTrackInfo(&r, "Title", data.Title)
	addTrackInfo(&r, "Artists", data.Artist)
	addTrackInfo(&r, "Album", data.Album)
	// addTrackInfo(&r, "Genres", data.Genre)
	addTrackInfo(&r, "Playtime", fmt.Sprintf("%s / %s", asTimecode(allData.CurrentElapsed), asTimecode(data.Length)))
	addMPDBar(&r)
	// addBar(&r, allData.CurrentElapsed/data.Length)
	return r
}

func formatUpcomingSongs(allData *domain.MpdState) string {
	songsRemaining := len(allData.Tracks) - (int(allData.CurrentIdx) + 1)
	r := fmt.Sprintf("%s${voffset -9}${hr}\n${alignr}. . . %d   S O N G S   L E F T\n", c(0), songsRemaining)
	startIdx := int(allData.CurrentIdx) + 1
	endIdx := startIdx + 37
	if endIdx > len(allData.Tracks) {
		endIdx = len(allData.Tracks)
	}
	upcomingTracks := allData.Tracks[startIdx:endIdx]
	for n, track := range upcomingTracks {
		r += fmt.Sprintf("%s%d. %s", c(0), n+1, formatSongMinimal(&track))
	}
	return r
}

func formatSongMinimal(track *domain.Track) string {
	return fmt.Sprintf("%s%s %sby %s%s", c(1), track.Title, c(0), c(2), track.Artist) + "\n"
}

func hasData(data *domain.MpdState) bool {
	return len(data.Tracks) > 0
}

func addArt(to *string) {
	*to += "${image /tmp/mpd-albumart.jpg -p 0,105 -s 200x200}${voffset 190}\n"
}

func addTrackInfo(to *string, name string, data string) {

	if data == "" {
		return
	}
	*to += fmt.Sprintf("%s%s:${alignr}%s%s\n",
		c(1), name, c(2), scroll(data, 20))
}

func addBar(to *string, percent float64) {
	if percent < 0.0 {
		percent = 0.0
	} else if percent > 1.0 {
		percent = 1.0
	}
	val := percent * 100
	*to += fmt.Sprintf("%s${execbar expr %.2f}\n", c(2), val)
}

func addMPDBar(to *string) {
	*to += fmt.Sprintf("%s${mpd_bar}\n", c(2))
}

func asTimecode(val float64) string {
	intVal := int64(val)
	mins := intVal / 60
	secs := intVal % 60
	return fmt.Sprintf("%01d:%02d", mins, secs)
}

func scroll(val string, l int) string {
	if rlen(val) <= l {
		return val
	}
	diff := rlen(val) - l
	unix := time.Now().Unix()
	step := int(unix % int64(diff))
	return subs(val, step, l+step+1)
}

func etcetera(val string, l int) string {
	if rlen(val) <= l {
		return val
	}
	return subs(val, 0, l-3) + "..."
}

// Change the color.
func c(n int) string {
	return fmt.Sprintf("${color%d}", n)
}

func max(val string, n int) string {
	if rlen(val) <= n {
		return val
	}
	return subs(val, 0, n)
}

// Allows to get a substring correctly (dealing with unicode runes).
func subs(s string, start int, end int) string {
	r := []rune(s)
	return string(r[start:end])
}

// Gets the rune length of a string.
func rlen(s string) int {
	return len([]rune(s))
}
