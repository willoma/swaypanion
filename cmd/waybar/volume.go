package main

import (
	socketclient "github.com/willoma/swaypanion/socket/client"
)

func volume(client *socketclient.Client) {
	conf := readConfig().Volume.Waybar

	receive(client, "volume subscribe", intPrinter(conf))
}
