package main

import (
	"flag"
	"os"
	"slices"
	"strings"

	"github.com/willoma/swaypanion/common"
	"github.com/willoma/swaypanion/waybar"
)

func main() {
	flag.String("c", ".config/swaypanion/waybar.conf", "Path to the swaypanion-waybar configuration (default .config/swaypanion/waybar.conf)")
	flag.Parse()

	eventTypes := waybar.EventTypes()

	args := flag.Args()
	if len(args) == 0 {
		common.LogError(
			"Event type missing. Need one of the following event types as an argument:\n\n"+
				strings.Join(eventTypes, "\n"),
			nil,
		)

		os.Exit(1)
	}

	if !slices.Contains(eventTypes, args[0]) {
		common.LogError(
			"Unknown event type. Need one of the following event types as an argument:\n\n"+
				strings.Join(eventTypes, "\n"),
			nil,
		)

		os.Exit(1)
	}

	if err := waybar.Subscribe(os.Stdout, args[0]); err != nil {
		common.LogError("Failed to subscribe", err)
		os.Exit(1)
	}
}
