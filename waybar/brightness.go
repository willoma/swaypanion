package waybar

import (
	"errors"
	"io"

	"github.com/willoma/swaypanion/socket"
	socketclient "github.com/willoma/swaypanion/socket/client"
)

func brightness(w io.Writer, client *socketclient.Client, conf *config) error {
	if err := client.Send(&socket.Message{
		Command: "brightness subscribe",
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

		alt, text, tooltip, disabled := conf.Backlight.formatValue(msg.Value)
		writeJSON(w, alt, text, tooltip, disabled)
	}
}
