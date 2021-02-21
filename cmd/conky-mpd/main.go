package main

import (
	"fmt"
	"os"

	"github.com/liampulles/bunsen-tools/cmd/conky-mpd/internal/format"
	"github.com/liampulles/bunsen-tools/cmd/conky-mpd/internal/mpd"
)

func main() {
	data, err := mpd.GetMpdState()
	if err != nil {
		// Quit silently
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err.Error())
		os.Exit(1)
	}
	formatted := format.Format(data)
	fmt.Print(formatted)
	os.Exit(0)
}
