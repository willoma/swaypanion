package modules

import (
	"errors"

	"github.com/willoma/swaypanion/common"
	"github.com/willoma/swaypanion/config"
	"github.com/willoma/swaypanion/socket"
	socketserver "github.com/willoma/swaypanion/socket/server"
)

const playerLabel = "player"

func (p *Player) SocketCommands() socketserver.Commands {
	return socketserver.NewCommands(
		p.socketGet, playerLabel, "Get current player status",
		p.socketGet, playerLabel+" get", "Get current player status",
		p.socketPlaypause, playerLabel+" playpause", "Play or pause reading",
		p.socketNext, playerLabel+" next", "Skip to next track",
		p.socketPrevious, playerLabel+" prev", "Reverse to previous track",
		p.socketSubscribe, playerLabel+" subscribe", "Get player status each time it changes",
		p.socketUnsubscribe, playerLabel+" unsubscribe", "Stop getting player status on change",
		p.socketShow, playerLabel+" show", "Show or start the player window",
	)
}

func (p *Player) socketGet(conn *socketserver.Connection, _ string, _ []string) {
	data, ok := p.get()
	if !ok {
		conn.SendError("failed to read player status")
		return
	}

	conn.Send(socket.Message{
		Command:    playerLabel,
		Value:      data.Status,
		Complement: data.complement(),
	})
}

func (p *Player) socketPlaypause(conn *socketserver.Connection, _ string, _ []string) {
	if err := p.playpause(); err != nil {
		if errors.Is(err, ErrNoPlayer) {
			conn.SendError(err.Error())
		} else {
			common.LogError("Failed to play or pause", err)
			conn.SendError("failed to play or pause")
		}
	}
}

func (p *Player) socketPrevious(conn *socketserver.Connection, _ string, _ []string) {
	if err := p.previous(); err != nil {
		if errors.Is(err, ErrNoPlayer) {
			conn.SendError(err.Error())
		} else {
			common.LogError("Failed to reverse to previous track", err)
			conn.SendError("failed to reverse to previous track")
		}
	}
}

func (p *Player) socketNext(conn *socketserver.Connection, _ string, _ []string) {
	if err := p.next(); err != nil {
		if errors.Is(err, ErrNoPlayer) {
			conn.SendError(err.Error())
		} else {
			common.LogError("Failed to skip to next track", err)
			conn.SendError("failed to skip to next track")
		}
	}
}

func (p *Player) socketSubscribe(conn *socketserver.Connection, _ string, _ []string) {
	p.subscriptions.Subscribe(conn, true, func(value playerData) {
		conn.Send(socket.Message{
			Command:    playerLabel,
			Value:      value.Status,
			Complement: value.complement(),
		})
	})
}

func (p *Player) socketUnsubscribe(conn *socketserver.Connection, _ string, _ []string) {
	p.subscriptions.Unsubscribe(conn)
}

func (p *Player) socketShow(conn *socketserver.Connection, _ string, _ []string) {
	if err := p.startOrShow(); err != nil {
		if errors.Is(err, config.ErrNoCommand) {
			conn.SendError(err.Error())
			return
		}

		common.LogError("Failed to show player window", err)
		conn.SendError("failed to show player window")
	}
}
