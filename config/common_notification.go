package config

import "time"

type NotificationSectionMessage struct {
	Enabled *bool         `yaml:"enabled"`
	Timeout time.Duration `yaml:"timeout"`
}

func (n NotificationSectionMessage) applyDefault(def NotificationSectionMessage) NotificationSectionMessage {
	if n.Enabled == nil {
		n.Enabled = def.Enabled
	}

	if n.Timeout == 0 {
		n.Timeout = def.Timeout
	}

	return n
}

type NotificationSectionPercent struct {
	Enabled *bool         `yaml:"enabled"`
	Timeout time.Duration `yaml:"timeout"`
	Format0 string        `yaml:"format0"`
	Formats []string      `yaml:"formats"`
}

func (n NotificationSectionPercent) applyDefault(def NotificationSectionPercent) NotificationSectionPercent {
	if n.Enabled == nil {
		n.Enabled = def.Enabled
	}

	if n.Timeout == 0 {
		n.Timeout = def.Timeout
	}

	if n.Format0 == "" {
		n.Format0 = def.Format0
	}

	if len(n.Formats) == 0 {
		n.Formats = make([]string, len(def.Formats))
		copy(n.Formats, def.Formats)
	}

	return n
}
