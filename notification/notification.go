package notification

import (
	"github.com/godbus/dbus/v5"
	"github.com/willoma/swaypanion/common"
)

type Notification struct {
	dbus *dbus.Conn
}

func New() (*Notification, error) {
	dbusConn, err := dbus.SessionBus()
	if err != nil {
		return nil, common.Errorf("failed to connect to DBus", err)
	}

	return &Notification{
		dbus: dbusConn,
	}, nil
}
