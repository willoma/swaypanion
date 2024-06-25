package swaypanion

import (
	"log/slog"
	"strconv"
	"strings"

	"github.com/godbus/dbus/v5"
)

const (
	defaultNotificationLevelTimeout              = 2000
	defaultNotificationLevelLabelBrightness      = " "
	defaultNotificationLevelLabelVolumeMuted     = " "
	defaultNotificationLevelLabelVolumeLow       = " "
	defaultNotificationLevelLabelVolumeHigh      = " "
	defaultNotificationLevelSummaryFormat        = "<label> <value> %"
	defaultNotificationLevelSummaryNoValueFormat = "<label>"
)

type NotificationConfig struct {
	Disable bool `yaml:"disable"`
	Level   struct {
		Timeout              int               `yaml:"timeout"`
		Labels               map[string]string `yaml:"labels"`
		SummaryFormat        string            `yaml:"summary_format"`
		SummaryNoValueFormat string            `yaml:"summary_no_value_format"`
	} `yaml:"timeout"`
}

func (c NotificationConfig) withDefaults() NotificationConfig {
	if c.Level.Timeout == 0 {
		c.Level.Timeout = defaultNotificationLevelTimeout
	}

	if c.Level.Labels == nil {
		c.Level.Labels = map[string]string{}
	}

	if c.Level.Labels["brightness"] == "" {
		c.Level.Labels["brightness"] = defaultNotificationLevelLabelBrightness
	}

	if c.Level.Labels["volume_muted"] == "" {
		c.Level.Labels["volume_muted"] = defaultNotificationLevelLabelVolumeMuted
	}

	if c.Level.Labels["volume_low"] == "" {
		c.Level.Labels["volume_low"] = defaultNotificationLevelLabelVolumeLow
	}

	if c.Level.Labels["volume_high"] == "" {
		c.Level.Labels["volume_high"] = defaultNotificationLevelLabelVolumeHigh
	}

	if c.Level.SummaryFormat == "" {
		c.Level.SummaryFormat = defaultNotificationLevelSummaryFormat
	}

	if c.Level.SummaryNoValueFormat == "" {
		c.Level.SummaryNoValueFormat = defaultNotificationLevelSummaryNoValueFormat
	}

	return c
}

type Notification struct {
	conf    NotificationConfig
	levelID uint32

	dbus *dbus.Conn
}

func NewNotification(conf NotificationConfig) (*Notification, error) {
	if conf.Disable {
		return nil, nil
	}

	dbusConn, err := dbus.SessionBus()
	if err != nil {
		return nil, err
	}

	return &Notification{
		conf: conf,
		dbus: dbusConn,
	}, nil

}

func (n *Notification) NotifyLevel(labelName string, value int) {
	var (
		summary string
		hints   = map[string]dbus.Variant{}
	)

	if value == -1 {
		summary = strings.ReplaceAll(n.conf.Level.SummaryNoValueFormat, "<label>", n.conf.Level.Labels[labelName])
	} else {
		summary = strings.ReplaceAll(
			strings.ReplaceAll(n.conf.Level.SummaryFormat, "<label>", n.conf.Level.Labels[labelName]),
			"<value>",
			strconv.Itoa(value),
		)
		hints["value"] = dbus.MakeVariant(value)
	}
	call := n.dbus.Object(
		"org.freedesktop.Notifications",
		"/org/freedesktop/Notifications",
	).Call(
		"org.freedesktop.Notifications.Notify", 0,
		"swaypanion-level",   // app_name
		n.levelID,            // replaces_id
		"",                   // app_icon
		summary,              // summary
		"",                   // body
		[]string{},           // actions
		hints,                // hints
		n.conf.Level.Timeout, // expire_timeout
	)
	if call.Err != nil {
		slog.Error("Failed to send level notification", "error", call.Err)
	}

	if len(call.Body) > 0 {
		if id, ok := call.Body[0].(uint32); ok {
			n.levelID = id
		}
	}
}
