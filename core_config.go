package swaypanion

import (
	"bytes"
	"io"
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
	Backlight    BacklightConfig    `yaml:"backlight"`
	Player       PlayerConfig       `yaml:"player"`
	Volume       VolumeConfig       `yaml:"volume"`
	Windows      WindowsConfig      `yaml:"windows"`
	Workspaces   WorkspacesConfig   `yaml:"workspaces"`
}

func (c *config) haveDefaults() {
	c.Notification = c.Notification.withDefaults()
	c.Backlight = c.Backlight.withDefaults()
	c.Player = c.Player.withDefaults()
	c.Volume = c.Volume.withDefaults()
	c.Windows = c.Windows.withDefaults()
	c.Workspaces = c.Workspaces.withDefaults()
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

func (s *Swaypanion) dumpConfig(w io.Writer) error {
	var buf bytes.Buffer

	buf.WriteString("Current configuration\n---------------------\n\n")

	if err := yaml.NewEncoder(&buf).Encode(s.conf); err != nil {
		return err
	}

	_, err := buf.WriteTo(w)
	return err
}

func (s *Swaypanion) dumpDefaultConfig(w io.Writer) error {
	conf := &config{}
	conf.haveDefaults()

	var buf bytes.Buffer

	buf.WriteString("Default configuration\n---------------------\n\n")

	if err := yaml.NewEncoder(&buf).Encode(conf); err != nil {
		return err
	}

	_, err := buf.WriteTo(w)
	return err
}
