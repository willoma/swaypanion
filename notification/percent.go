package notification

import (
	"strconv"
	"strings"
	"sync"

	"github.com/godbus/dbus/v5"

	"github.com/willoma/swaypanion/common"
	"github.com/willoma/swaypanion/config"
)

type PercentNotifier struct {
	notification *Notification

	disabled bool

	timeout        int
	formatDisabled string
	format0        string
	formats        []string
	formatStepSize int
	format100      string

	mu             sync.Mutex
	notificationID uint32
}

func (n *Notification) PercentNotifier() *PercentNotifier {
	return &PercentNotifier{
		notification: n,
	}
}

func (p *PercentNotifier) Reconfigure(conf config.NotificationSectionPercent) {
	if conf.Enabled == nil || !*conf.Enabled {
		p.disabled = true
		return
	}

	p.disabled = false
	p.timeout = int(conf.Timeout.Milliseconds())
	p.formatDisabled = conf.FormatDisabled
	p.format0 = conf.Format0
	p.formats = conf.Formats

	if len(conf.Formats) == 0 {
		p.formatStepSize = 100
		p.formats = []string{""}
		p.format100 = ""
	} else {
		p.formatStepSize = 100 / len(conf.Formats)
		p.format100 = conf.Formats[len(conf.Formats)-1]
	}
}

func (p *PercentNotifier) Notify(percent common.Int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.disabled {
		return
	}

	var format string

	if percent.Disabled {
		format = p.formatDisabled
	} else if percent.Value <= 0 {
		format = p.format0
	} else if percent.Value >= 100 {
		format = p.format100
	} else {
		format = p.formats[percent.Value/p.formatStepSize]
	}

	content := strings.ReplaceAll(format, "{value}", strconv.Itoa(percent.Value))

	hints := map[string]dbus.Variant{"value": dbus.MakeVariant(percent.Value)}

	call := p.notification.dbus.Object(
		"org.freedesktop.Notifications",
		"/org/freedesktop/Notifications",
	).Call(
		"org.freedesktop.Notifications.Notify", 0,
		"swaypanion-level", // app_name
		p.notificationID,   // replaces_id
		"",                 // app_icon
		content,            // summary
		"",                 // body
		[]string{},         // actions
		hints,              // hints
		p.timeout,          // expire_timeout
	)
	if call.Err != nil {
		common.LogError("Failed to send level notification", call.Err)
	}

	if len(call.Body) > 0 {
		if id, ok := call.Body[0].(uint32); ok {
			p.notificationID = id
		}
	}
}
