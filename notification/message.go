package notification

import (
	"github.com/godbus/dbus/v5"

	"github.com/willoma/swaypanion/common"
	"github.com/willoma/swaypanion/config"
)

type MessageNotifier struct {
	notification *Notification

	timeout int
}

func (n *Notification) MessageNotifier() *MessageNotifier {
	return &MessageNotifier{
		notification: n,
	}
}

func (m *MessageNotifier) Reconfigure(config config.NotificationSectionMessage) {
	m.timeout = int(config.Timeout.Milliseconds())
}

func (m *MessageNotifier) Notify(message string) {
	call := m.notification.dbus.Object(
		"org.freedesktop.Notifications",
		"/org/freedesktop/Notifications",
	).Call(
		"org.freedesktop.Notifications.Notify", 0,
		"swaypanion-message",      // app_name
		uint32(0),                 // replaces_id
		"",                        // app_icon
		message,                   // summary
		"",                        // body
		[]string{},                // actions
		map[string]dbus.Variant{}, // hints
		m.timeout,                 // expire_timeout
	)
	if call.Err != nil {
		common.LogError("Failed to send message notification", call.Err)
	}
}
