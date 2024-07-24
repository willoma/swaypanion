package main

import (
	"errors"
	"io"
	"strings"

	"github.com/willoma/swaypanion/socket"
	socketclient "github.com/willoma/swaypanion/socket/client"
)

func runDirect(client *socketclient.Client, commands []string) {
	go func() {
		for _, c := range commands {
			fields := strings.Split(c, string(cliFieldSeparator))

			msg := &socket.Message{}

			switch len(fields) {
			case 0:
				continue
			case 1:
				msg.Command = fields[0]
			case 2:
				msg.Command = fields[0]
				msg.Value = fields[1]
			default:
				msg.Command = fields[0]
				msg.Value = fields[1]
				msg.Complement = fields[2:]
			}

			if err := client.Send(msg); err != nil {
				printError("Failed to send command", err)
			}
		}

		if err := client.Send(&socket.Message{
			Command: "close",
		}); err != nil {
			printError("Failed to send close command", err)
		}
	}()

	for {
		msg, err := client.Read()
		if err != nil {
			if !errors.Is(err, io.EOF) {
				printError("Failed to read response", err)
			}

			return
		}

		printResponse(msg)
	}
}
