package main

import (
	"strings"

	"github.com/willoma/swaypanion/socket"
	socketclient "github.com/willoma/swaypanion/socket/client"
)

func player(client *socketclient.Client) {
	conf := readConfig().Player.Waybar

	receive(client, "player subscribe", func(msg *socket.Message) {
		replaces := map[string]string{
			"status":  msg.Value,
			"artist":  "",
			"artists": "",
			"album":   "",
			"title":   "",
		}

		for _, c := range msg.Complement {
			splat := strings.SplitN(c, ":", 2)
			if len(splat) != 2 {
				continue
			}

			value := strings.TrimSpace(splat[1])

			switch splat[0] {
			case "Artist":
				replaces["artist"] = value
			case "Artists":
				replaces["artists"] = value
			case "Album":
				replaces["album"] = value
			case "Title":
				replaces["title"] = value
			}
		}

		if replaces["artists"] == "" {
			replaces["artists"] = replaces["artist"]
		}

		printJSON(conf.FormatValue(msg.Value, replaces))
	})
}
