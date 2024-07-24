package main

import (
	socketclient "github.com/willoma/swaypanion/socket/client"
)

func brightness(client *socketclient.Client) {
	conf := readConfig().Backlight.Waybar

	receive(client, "brightness subscribe", intPrinter(conf))
}
