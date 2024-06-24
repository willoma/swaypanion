package swaypanion

import (
	"os"
	"strings"

	"github.com/adrg/xdg"
	"gopkg.in/yaml.v3"
)

type CommandType string

const (
	CommandTypeExec CommandType = "exec"
	CommandTypeSway CommandType = "sway"

	defaultCommandType = CommandTypeSway
)

func (c CommandType) defaultIfInvalid() CommandType {
	if c == CommandTypeExec || c == CommandTypeSway {
		return c
	}

	return defaultCommandType
}

type config struct {
	Notification NotificationConfig `yaml:"notification"`
	Workspaces   WorkspacesConfig   `yaml:"workspaces"`
	Backlight    BacklightConfig    `yaml:"backlight"`
	Player       PlayerConfig       `yaml:"player"`
	Volume       VolumeConfig       `yaml:"volume"`
}

func (c *config) haveDefaults() {
	c.Notification = c.Notification.withDefaults()
	c.Workspaces = c.Workspaces.withDefaults()
	c.Backlight = c.Backlight.withDefaults()
	c.Player = c.Player.withDefaults()
	c.Volume = c.Volume.withDefaults()
}

func newConfig() (*config, error) {
	confFileName, err := xdg.SearchConfigFile("swaypanion.yaml")
	if err != nil {
		if strings.Contains(err.Error(), "could not locate") {
			conf := &config{}
			conf.haveDefaults()
			return conf, nil
		}

		return nil, err
	}

	confFile, err := os.Open(confFileName)
	if err != nil {
		return nil, err
	}
	defer confFile.Close()

	conf := &config{}
	if err := yaml.NewDecoder(confFile).Decode(conf); err != nil {
		return nil, err
	}

	conf.haveDefaults()

	return conf, nil
}
