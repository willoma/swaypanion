package swaypanion

import (
	socketserver "github.com/willoma/swaypanion/socket/server"
)

type socketCommander interface {
	SocketCommands() socketserver.Commands
}

type Stopper interface {
	Stop()
}
