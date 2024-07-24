package modules

import (
	"errors"
	"net"
	"strconv"

	"github.com/willoma/swaypanion/common"
	socketserver "github.com/willoma/swaypanion/socket/server"
)

const volumeLabel = "volume"

func (v *Volume) SocketCommands() socketserver.Commands {
	return socketserver.NewCommands(
		v.socketGet, volumeLabel, "Get current volume",
		v.socketGet, volumeLabel+" get", "Get current volume",
		v.socketUp, volumeLabel+" up", "Increase volume",
		v.socketDown, volumeLabel+" down", "Decrease volume",
		v.socketSet, volumeLabel+" set", "Set volume", "volume percent",
		v.socketSubscribe, volumeLabel+" subscribe", "Get volume each time it changes",
		v.socketUnsubscribe, volumeLabel+" unsubscribe", "Stop getting volume on change",
	)
}

func (v *Volume) socketGet(conn *socketserver.Connection, _ string, _ []string) {
	value, ok := v.get()
	if !ok {
		conn.SendError("failed to read volume")
		return
	}

	if err := conn.SendInt(volumeLabel, value); err != nil {
		common.LogError("Failed to send volume", err)
	}
}

func (v *Volume) socketSet(conn *socketserver.Connection, value string, _ []string) {
	if value == "" {
		conn.SendError("missing volume value")
		return
	}

	percent, err := strconv.Atoi(value)
	if err != nil {
		conn.SendError("failed to convert argument to int")
	}

	newValue, ok := v.set(percent)
	if !ok {
		conn.SendError("failed to change volume")
		return
	}

	if err := conn.SendInt(volumeLabel, newValue); err != nil {
		common.LogError("Failed to send volume", err)
	}
}

func (v *Volume) socketUp(conn *socketserver.Connection, _ string, _ []string) {
	value, ok := v.up()
	if !ok {
		conn.SendError("failed to change volume")
		return
	}

	if err := conn.SendInt(volumeLabel, value); err != nil {
		common.LogError("Failed to send volume", err)
	}
}

func (v *Volume) socketDown(conn *socketserver.Connection, _ string, _ []string) {
	value, ok := v.down()
	if !ok {
		conn.SendError("failed to change volume")
		return
	}

	if err := conn.SendInt(volumeLabel, value); err != nil {
		common.LogError("Failed to send volume", err)
	}
}

func (v *Volume) socketSubscribe(conn *socketserver.Connection, _ string, _ []string) {
	v.subscriptions.Subscribe(conn, true, func(value common.Int) {
		var err error

		if value.Value == -1 {
			err = conn.SendString(volumeLabel, "mute")
		} else {
			err = conn.SendInt(volumeLabel, value.Value)
		}

		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				v.subscriptions.Unsubscribe(conn)
				return
			}

			common.LogError("Failed to send subscribed volume", err)
		}
	})
}

func (v *Volume) socketUnsubscribe(conn *socketserver.Connection, _ string, _ []string) {
	v.subscriptions.Unsubscribe(conn)
}
