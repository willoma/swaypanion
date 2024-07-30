package waybar

import (
	"errors"
	"io"
	"strings"

	"github.com/willoma/swaypanion/socket"
	socketclient "github.com/willoma/swaypanion/socket/client"
)

func player(w io.Writer, client *socketclient.Client, conf *config) error {
	if err := client.Send(&socket.Message{
		Command: "player subscribe",
	}); err != nil {
		return err
	}

	for {
		msg, err := client.Read()
		if err != nil {
			if !errors.Is(err, io.EOF) {
				return err
			}

			return nil
		}

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

		alt, text, tooltip, disabled := conf.Player.formatValue(msg.Value, replaces)
		writeJSON(w, alt, text, tooltip, disabled)
	}
}
