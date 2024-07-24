package main

import (
	"os"

	socketclient "github.com/willoma/swaypanion/socket/client"
)

const (
	carriageReturn byte = 13
)

func main() {
	client, err := socketclient.New()
	if err != nil {
		printError("Failed to run socket client", err)
	}

	defer func() {
		if err := client.Close(); err != nil {
			printError("Failed to close connection to server", err)
		}
	}()

	if len(os.Args) == 1 {
		runInteractive(client)
		return
	}

	runDirect(client, os.Args[1:])
}
