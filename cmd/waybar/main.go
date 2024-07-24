package main

import (
	"os"

	"github.com/willoma/swaypanion/common"
	"github.com/willoma/swaypanion/config"
	socketclient "github.com/willoma/swaypanion/socket/client"
)

func main() {
	if len(os.Args) == 1 {
		showEventTypes()
	}

	client, err := socketclient.New()
	if err != nil {
		printError("Failed to run socket client", err)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "brightness":
		brightness(client)
	case "player":
		player(client)
	case "volume":
		volume(client)
	default:
		showEventTypes()
	}

	defer func() {
		if err := client.Close(); err != nil {
			printError("Failed to close connection to server", err)
		}
	}()
}

func readConfig() *config.Config {
	conf, err := config.New(nil)
	if err != nil {
		printError("Failed to read configuration", err)
	}

	return conf
}

func showEventTypes() {
	common.LogError(`Need one of the following event types as an argument:

brightness
player
volume
`, nil)

	os.Exit(1)
}
