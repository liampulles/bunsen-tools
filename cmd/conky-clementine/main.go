package main

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/godbus/dbus/v5"
)

func main() {
	data, err := getData()
	if err != nil {
		fmt.Println(err)
	}
	formatted := format(data)
	fmt.Print(formatted)
	os.Exit(0)
}

type clementineState struct {
	CurrentTrack currentTrack
}

type currentTrack struct {
	Title       string
	Album       string
	Artists     []string
	Genres      []string
	TrackNumber int32
	DiscNumber  int32
	Year        int32

	Position int64
	Length   int64

	ArtURL string
}

func format(data *clementineState) string {
	r := "\nC L E M E N T I N E\n"
	addArt(&r, data)
	addTrackInfo(&r, "Title", data.CurrentTrack.Title)
	addTrackInfo(&r, "Artists", strings.Join(data.CurrentTrack.Artists, " "))
	addTrackInfo(&r, "Album", data.CurrentTrack.Album)
	addTrackInfo(&r, "Genres", strings.Join(data.CurrentTrack.Genres, " "))
	addTrackInfo(&r, "Playtime", fmt.Sprintf("%s / %s", asTimecode(data.CurrentTrack.Position), asTimecode(data.CurrentTrack.Length)))
	return r
}

func addArt(to *string, data *clementineState) {
	if data.CurrentTrack.ArtURL == "" {
		return
	}
	*to += fmt.Sprintf("${hr}${image %s -p 0,135 -s 200x200}${voffset 200}\n",
		asPath(data.CurrentTrack.ArtURL))
}

func addTrackInfo(to *string, name string, data string) {
	if data == "" {
		return
	}
	*to += fmt.Sprintf("%s:${alignr}%s\n",
		name, scroll(data, 20))
}

func getData() (*clementineState, error) {
	conn, err := dbus.SessionBus()
	if err != nil {
		return nil, fmt.Errorf("session connect error: %w", err)
	}
	defer conn.Close()

	currentTrack, err := getCurrentTrack(conn)
	if err != nil {
		return nil, fmt.Errorf("could not get current track: %w", err)
	}

	return &clementineState{
		CurrentTrack: *currentTrack,
	}, nil
}

func getCurrentTrack(conn *dbus.Conn) (*currentTrack, error) {
	var metadata map[string]dbus.Variant
	if err := getPlayerProperty(conn, "Metadata", &metadata); err != nil {
		return nil, fmt.Errorf("get metadata property error: %w", err)
	}

	var position int64
	if err := getPlayerProperty(conn, "Position", &position); err != nil {
		return nil, fmt.Errorf("get metadata property error: %w", err)
	}

	return &currentTrack{
		Title:       safeAsString(metadata["xesam:title"].Value()),
		Album:       safeAsString(metadata["xesam:album"].Value()),
		Artists:     safeAsStringSlice(metadata["xesam:artist"].Value()),
		Genres:      safeAsStringSlice(metadata["xesam:genre"].Value()),
		TrackNumber: safeAsInt32(metadata["xesam:trackNumber"].Value()),
		DiscNumber:  safeAsInt32(metadata["xesam:discNumber"].Value()),
		Year:        safeAsInt32(metadata["year"].Value()),

		Position: position,
		Length:   safeAsInt64(metadata["mpris:length"].Value()),

		ArtURL: safeAsString(metadata["mpris:artUrl"].Value()),
	}, nil
}

func getPlayerProperty(conn *dbus.Conn, name string, ptr interface{}) error {
	return conn.Object("org.mpris.MediaPlayer2.clementine", "/org/mpris/MediaPlayer2").Call("org.freedesktop.DBus.Properties.Get", 0, "org.mpris.MediaPlayer2.Player", name).Store(ptr)
}

func asTimecode(val int64) string {
	val = val / 1000000
	mins := val / 60
	secs := val % 60
	return fmt.Sprintf("%01d:%02d", mins, secs)
}

func scroll(val string, l int64) string {
	if int64(len(val)) < l {
		return val
	}
	repeated := val + " | "
	repeated = repeated + repeated
	repeated = repeated + repeated
	unix := time.Now().Unix()
	step := unix % int64(len(repeated)/2)
	return repeated[step : l+step]
}

func asPath(fileURL string) string {
	p, _ := url.Parse(fileURL)
	return p.Path
}

func safeAsStringSlice(in interface{}) []string {
	if in == nil {
		return nil
	}
	return in.([]string)
}

func safeAsString(in interface{}) string {
	if in == nil {
		return ""
	}
	return in.(string)
}

func safeAsInt32(in interface{}) int32 {
	if in == nil {
		return 0
	}
	return in.(int32)
}

func safeAsInt64(in interface{}) int64 {
	if in == nil {
		return 0
	}
	return in.(int64)
}
