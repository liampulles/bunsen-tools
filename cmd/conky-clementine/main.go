package main

import (
	"fmt"
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
	fmt.Println(formatted)
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
	return fmt.Sprintf(`
C L E M E N T I N E
${hr}
Title:${alignr}%s
Artists:${alignr}%s
Album:${alignr}%s
Genre:${alignr}%s
Playtime:${alignr}%s / %s`,
		scroll(data.CurrentTrack.Title, 20),
		scroll(strings.Join(data.CurrentTrack.Artists, " "), 20),
		scroll(data.CurrentTrack.Album, 20),
		scroll(strings.Join(data.CurrentTrack.Genres, " "), 20),
		asTimecode(data.CurrentTrack.Position), asTimecode(data.CurrentTrack.Length),
	)
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
		Title:       metadata["xesam:title"].Value().(string),
		Album:       metadata["xesam:album"].Value().(string),
		Artists:     metadata["xesam:artist"].Value().([]string),
		Genres:      metadata["xesam:genre"].Value().([]string),
		TrackNumber: metadata["xesam:trackNumber"].Value().(int32),
		DiscNumber:  metadata["xesam:discNumber"].Value().(int32),
		Year:        metadata["year"].Value().(int32),

		Position: position,
		Length:   metadata["mpris:length"].Value().(int64),

		ArtURL: metadata["mpris:artUrl"].Value().(string),
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
	repeated := val + "  "
	repeated = repeated + repeated
	repeated = repeated + repeated
	unix := time.Now().Unix()
	step := unix % l
	return repeated[step : l+step]
}
