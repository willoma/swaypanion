package waybar

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const systemconfPath = "/etc/swaypanion/waybar.conf"

type config struct {
	Backlight configPercent `yaml:"backlight"`
	Player    configString  `yaml:"player"`
	Volume    configPercent `yaml:"volume"`
}

var defaultConfig = &config{
	Backlight: configPercent{
		Icon0:       "",
		Icons:       []string{"", "", ""},
		TextFormat0: " {value} %",
		TextFormats: []string{
			" {value} %",
			" {value} %",
			" {value} %",
		},
		TooltipFormat0: "",
		TooltipFormats: []string{""},
	},
	Player: configString{
		Icons: map[string]string{
			"no player found": "",
			"Playing":         "",
			"Paused":          "",
			"Stopped":         "",
		},
		FormatText:    "{icon}{artist? }{artist}{title? - }{title}",
		FormatTooltip: "{artists}{album?\n }{album}{title?\n}{title}",
	},
	Volume: configPercent{
		IconDisabled:       "",
		Icon0:              "",
		Icons:              []string{"", ""},
		TextFormatDisabled: "",
		TextFormat0:        " {value} %",
		TextFormats: []string{
			" {value} %",
			" {value} %",
		},
		TooltipFormat0: "",
		TooltipFormats: []string{""},
	},
}

func readConfig(configPath string) (*config, error) {
	if configPath != "" {
		if !filepath.IsAbs(configPath) {
			home, err := os.UserHomeDir()
			if err != nil {
				return nil, err
			}

			configPath = filepath.Join(home, configPath)
		}

		if uStat, uErr := os.Stat(configPath); uErr == nil && !uStat.IsDir() {
			return readConfigYAML(configPath)
		}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	userconfPath := filepath.Join(home, ".config", "swaypanion", "waybar.conf")

	if uStat, uErr := os.Stat(userconfPath); uErr == nil && !uStat.IsDir() {
		return readConfigYAML(userconfPath)
	}

	if sStat, sErr := os.Stat(systemconfPath); sErr == nil && !sStat.IsDir() {
		return readConfigYAML(systemconfPath)
	}

	c := &config{}
	c.applyDefault()

	return c, nil
}

func readConfigYAML(path string) (*config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	c := &config{}

	if err := yaml.NewDecoder(f).Decode(c); err != nil {
		return nil, err
	}

	c.applyDefault()

	return c, nil
}

func (c *config) applyDefault() {
	c.Backlight = c.Backlight.applyDefault(defaultConfig.Backlight)
	c.Player = c.Player.applyDefault(defaultConfig.Player)
	c.Volume = c.Volume.applyDefault(defaultConfig.Volume)
}
