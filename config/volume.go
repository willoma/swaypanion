package config

import "time"

type Volume struct {
	config[*Volume] `yaml:"-"`

	Server       string                     `yaml:"pulseaudio_server"`
	SinkName     string                     `yaml:"pulseaudio_sink_name"`
	StepSize     int                        `yaml:"pulseaudio_step_size"`
	Notification NotificationSectionPercent `yaml:"notification"`
}

var DefaultVolume = &Volume{
	Server:   "",
	SinkName: "@DEFAULT_SINK@",
	StepSize: 5,
	Notification: NotificationSectionPercent{
		Enabled:        &trueValue,
		Timeout:        2 * time.Second,
		FormatDisabled: "",
		Format0:        " {value} %",
		Formats: []string{
			" {value} %",
			" {value} %",
		},
	},
}

func (v *Volume) applyDefault() {
	if v.SinkName == "" {
		v.SinkName = DefaultVolume.SinkName
	}

	if v.StepSize == 0 {
		v.StepSize = DefaultVolume.StepSize
	}

	v.Notification = v.Notification.applyDefault(DefaultVolume.Notification)
}
