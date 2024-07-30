package waybar

import (
	"io"

	socketclient "github.com/willoma/swaypanion/socket/client"
)

func Subscribe(w io.Writer, eventType string) error {
	client, err := socketclient.New()
	if err != nil {
		return err
	}

	conf, err := readConfig()
	if err != nil {
		return err
	}

	switch eventType {
	case "brightness":
		brightness(w, client, conf)
	case "player":
		player(w, client, conf)
	case "volume":
		volume(w, client, conf)
	}

	return nil
}

func EventTypes() []string {
	return []string{
		"brightness",
		"player",
		"volume",
	}
}
