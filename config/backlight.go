package config

import (
	"os"
	"time"

	"github.com/willoma/swaypanion/common"
)

type Backlight struct {
	config[*Backlight] `yaml:"-"`

	DeviceName      string                     `yaml:"device_name"`
	MinimumPercent  float64                    `yaml:"minimum"`
	StepSizePercent float64                    `yaml:"step_size"`
	PollInterval    time.Duration              `yaml:"poll_interval"`
	Notification    NotificationSectionPercent `yaml:"notification"`
	Waybar          WaybarPercent              `yaml:"waybar"`
}

var DefaultBacklight = &Backlight{
	DeviceName:      "",
	MinimumPercent:  0.5,
	StepSizePercent: 5,
	PollInterval:    500 * time.Millisecond,
	Notification: NotificationSectionPercent{
		Enabled: &trueValue,
		Timeout: 2 * time.Second,
		Format0: " -",
		Formats: []string{
			"  {value} %",
			"  {value} %",
			"  {value} %",
		},
	},
	Waybar: WaybarPercent{
		Icon0:       "",
		Icons:       []string{"", "", ""},
		TextFormat0: " -",
		TextFormats: []string{
			"  {value} %",
			"  {value} %",
			"  {value} %",
		},
		TooltipFormat0: "",
		TooltipFormats: []string{""},
	},
}

func (b *Backlight) applyDefault() {
	if b.DeviceName == "" {
		b.findDeviceName()
	}

	if b.MinimumPercent == 0 {
		b.MinimumPercent = DefaultBacklight.MinimumPercent
	}

	if b.StepSizePercent == 0 {
		b.StepSizePercent = DefaultBacklight.StepSizePercent
	}

	if b.PollInterval == 0 {
		b.PollInterval = DefaultBacklight.PollInterval
	}

	b.Notification = b.Notification.applyDefault(DefaultBacklight.Notification)

	b.Waybar = b.Waybar.applyDefault(DefaultBacklight.Waybar)
}

func (b *Backlight) findDeviceName() {
	entries, err := os.ReadDir("/sys/class/backlight")
	if err != nil {
		common.LogError("Failed to look for default backlight device", err)
		return
	}

	if len(entries) == 0 {
		common.LogError("No backlight device available", nil)
		return
	}

	b.DeviceName = entries[0].Name()
}
