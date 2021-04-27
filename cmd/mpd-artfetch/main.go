package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"

	"github.com/fhs/gompd/v2/mpd"
)

func main() {
	if err := run(os.Args[1]); err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}

func run(libraryDir string) error {
	w, err := mpd.NewWatcher("tcp", ":6600", "", "player")
	if err != nil {
		return fmt.Errorf("watcher construct error: %w", err)
	}
	defer w.Close()

	// Do once to start off with.
	if err := saveArt(libraryDir); err != nil {
		fmt.Fprintf(os.Stderr, "save art error: %s\n", err.Error())
	}

	for range w.Event {
		if err := saveArt(libraryDir); err != nil {
			fmt.Fprintf(os.Stderr, "save art error: %s\n", err.Error())
		}
	}
	return nil
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

	if err := savePictureFFMPEG(uri); err != nil {
		return fmt.Errorf("save picture error: %w", err)
	}
	// bytes, err := readPicture(conn, uri)
	// if err != nil {
	// 	return fmt.Errorf("album art error: %w", err)
	// }

	// if err := ioutil.WriteFile("/tmp/mpd-albumart.jpg", bytes, 0644); err != nil {
	// 	return fmt.Errorf("file write error: %w", err)
	// }
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

func readPicture(conn *mpd.Client, uri string) ([]byte, error) {
	offset := 0
	var data []byte
	for {
		// Read the data in chunks
		chunk, size, err := conn.Command("readpicture %s %d", uri, offset).Binary()
		if err != nil {
			return nil, err
		}

		// Accumulate the data
		data = append(data, chunk...)
		offset = len(data)
		if offset >= size {
			break
		}
	}
	return data, nil
}
