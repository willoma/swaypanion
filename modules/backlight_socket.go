package modules

import (
	"errors"
	"net"
	"strconv"

	"github.com/willoma/swaypanion/common"
	socketserver "github.com/willoma/swaypanion/socket/server"
)

const brightnessLabel = "brightness"

func (b *Backlight) SocketCommands() socketserver.Commands {
	return socketserver.NewCommands(
		b.socketGet, brightnessLabel, "Get current brightness",
		b.socketGet, brightnessLabel+" get", "Get current brightness",
		b.socketUp, brightnessLabel+" up", "Increase brightness",
		b.socketDown, brightnessLabel+" down", "Decrease brightness",
		b.socketSet, brightnessLabel+" set", "Set brightness", "brightness percent",
		b.socketSubscribe, brightnessLabel+" subscribe", "Get brightness each time it changes",
		b.socketUnsubscribe, brightnessLabel+" unsubscribe", "Stop getting brightness on change",
	)
}

func (b *Backlight) socketGet(conn *socketserver.Connection, _ string, _ []string) {
	value, ok := b.get()
	if !ok {
		conn.SendError("failed to read brightness")
		return
	}

	if err := conn.SendInt(brightnessLabel, value); err != nil {
		common.LogError("Failed to send brightness", err)
	}
}

func (b *Backlight) socketSet(conn *socketserver.Connection, value string, _ []string) {
	if value == "" {
		conn.SendError("missing brightness value")
		return
	}

	percent, err := strconv.Atoi(value)
	if err != nil {
		conn.SendError("failed to convert argument to int")
	}

	newValue, ok := b.set(percent)
	if !ok {
		conn.SendError("failed to change brightness")
		return
	}

	if err := conn.SendInt(brightnessLabel, newValue); err != nil {
		common.LogError("Failed to send brightness", err)
	}
}

func (b *Backlight) socketUp(conn *socketserver.Connection, _ string, _ []string) {
	value, ok := b.up()
	if !ok {
		conn.SendError("failed to change brightness")
		return
	}

	if err := conn.SendInt(brightnessLabel, value); err != nil {
		common.LogError("Failed to send brightness", err)
	}
}

func (b *Backlight) socketDown(conn *socketserver.Connection, _ string, _ []string) {
	value, ok := b.down()
	if !ok {
		conn.SendError("failed to change brightness")
		return
	}

	if err := conn.SendInt(brightnessLabel, value); err != nil {
		common.LogError("Failed to send brightness", err)
	}
}

func (b *Backlight) socketSubscribe(conn *socketserver.Connection, _ string, _ []string) {
	b.subscriptions.Subscribe(conn, true, func(value common.Int) {
		if err := conn.SendInt(brightnessLabel, value.Value); err != nil {
			if errors.Is(err, net.ErrClosed) {
				b.subscriptions.Unsubscribe(conn)
				return
			}

			common.LogError("Failed to send subscribed brightness", err)

		}
	})
}

func (b *Backlight) socketUnsubscribe(conn *socketserver.Connection, _ string, _ []string) {
	b.subscriptions.Unsubscribe(conn)
}
