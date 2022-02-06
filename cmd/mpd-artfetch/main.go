package main

import (
	"fmt"
	"io"
	"log"
	"log/syslog"
	"os"
	"os/exec"
	"path"

	"github.com/fhs/gompd/v2/mpd"
)

func main() {
	setupLogger()
	run(os.Args[1])
}

func setupLogger() {
	sw, err := syslog.New(syslog.LOG_INFO, "mpd-artfetch")
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not establish syslog connection - failing")
		panic(err)
	}
	mw := io.MultiWriter(os.Stderr, sw)
	log.SetOutput(mw)
}

func run(libraryDir string) {
	w, err := mpd.NewWatcher("tcp", ":6600", "", "player")
	if err != nil {
		log.Fatalf("watcher construct error: %s", err.Error())
	}
	defer w.Close()

	// Do once to start off with.
	if err := saveArt(libraryDir); err != nil {
		log.Printf("save art error: %s\n", err.Error())
	}

	// Then run forever.
	for range w.Event {
		if err := saveArt(libraryDir); err != nil {
			log.Printf("save art error: %s\n", err.Error())
		}
	}
}

func saveArt(libraryDir string) error {
	conn, err := mpd.Dial("tcp", ":6600")
	if err != nil {
		return fmt.Errorf("dial error: %w", err)
	}
	defer conn.Close()

	attrs, err := conn.CurrentSong()
	if err != nil {
		return fmt.Errorf("current song error: %w", err)
	}
	uri := path.Join(libraryDir, attrs["file"])
	log.Printf("current song path: %s", uri)

	if err := savePictureFFMPEG(uri); err != nil {
		return fmt.Errorf("save picture error: %w", err)
	}
	return nil
}

func savePictureFFMPEG(uri string) error {
	cmd := exec.Command("ffmpeg", "-y", "-i", uri, "-an", "-vcodec", "copy", "/tmp/mpd-albumart.jpg")
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "ffmpeg error for %s - getting default...", uri)
		if err = copyDefault(); err != nil {
			return fmt.Errorf("copy default error: %w", err)
		}
	}
	return nil
}

func copyDefault() error {
	os.Remove("/tmp/mpd-albumart.jpg")
	source, err := os.Open("/usr/local/albumart.jpeg")
	if err != nil {
		return fmt.Errorf("open default error: %w", err)
	}
	defer source.Close()

	destination, err := os.Create("/tmp/mpd-albumart.jpg")
	if err != nil {
		return fmt.Errorf("open tmp albumart error: %w", err)
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	if err != nil {
		return fmt.Errorf("copy error: %w", err)
	}
	return nil
}
