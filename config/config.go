package config

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/willoma/swaypanion/common"
	"gopkg.in/yaml.v3"
)

type notifier interface {
	Notify(string)
}

const systemconfPath = "/etc/swaypanion.conf"

type Config struct {
	config[*Config] `yaml:"-"`

	Backlight *Backlight `yaml:"backlight"`
	Player    *Player    `yaml:"player"`
	Volume    *Volume    `yaml:"volume"`
	SwayNodes *SwayNodes `yaml:"sway"`

	CoreMessages NotificationSectionMessage `yaml:"core_messages"`

	notifier notifier          `yaml:"-"`
	watcher  *fsnotify.Watcher `yaml:"-"`

	userconfPath string `yaml:"-"`
	isUserConf   bool   `yaml:"-"`
}

var Default = &Config{
	Backlight: DefaultBacklight,
	Player:    DefaultPlayer,
	Volume:    DefaultVolume,
	SwayNodes: DefaultSwayNodes,
	CoreMessages: NotificationSectionMessage{
		Enabled: &trueValue,
		Timeout: 3 * time.Second,
	},
}

func New(notif notifier) (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	c := &Config{
		notifier:     notif,
		userconfPath: filepath.Join(home, ".config", "swaypanion.conf"),
	}

	if uStat, uErr := os.Stat(c.userconfPath); uErr == nil && !uStat.IsDir() {
		if c.readPath(c.userconfPath) {
			c.isUserConf = true
		}
	} else if sStat, sErr := os.Stat(systemconfPath); sErr == nil && !sStat.IsDir() {
		c.readPath(systemconfPath)
	} else {
		c.resetToDefault()
	}

	c.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	go c.watch()

	if err := c.watcher.Add(filepath.Join(home, ".config")); err != nil {
		return nil, err
	}

	if err := c.watcher.Add("/etc"); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Config) Close() error {
	return c.watcher.Close()
}

func (c *Config) readPath(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			common.LogError("Failed to open new configuration in "+c.userconfPath, err)
		}

		return false
	}

	defer f.Close()

	c.reset()

	if err := yaml.NewDecoder(f).Decode(c); err != nil {
		var yErr *yaml.TypeError
		if errors.As(err, &yErr) {
			for _, msg := range yErr.Errors {
				common.LogError("Failed to read configuration:"+msg, nil)
			}
		}
	}

	c.applyDefault()

	return true
}

func (c *Config) watch() {
	for {
		select {
		case event, ok := <-c.watcher.Events:
			if !ok {
				return
			}

			switch event.Name {
			case c.userconfPath:
				switch event.Op {
				case fsnotify.Create, fsnotify.Write:
					if c.readPath(c.userconfPath) {
						c.isUserConf = true
						c.announceReloaded()
					}
				case fsnotify.Remove:
					if stat, err := os.Stat(systemconfPath); err != nil || stat.IsDir() {
						c.resetToDefault()
						c.announceReloaded()
					}

					if c.readPath(systemconfPath) {
						c.isUserConf = false
						c.announceReloaded()
					} else {
						c.resetToDefault()
						c.announceReloaded()
					}
				}
			case systemconfPath:
				if c.isUserConf {
					continue
				}

				switch event.Op {
				case fsnotify.Create, fsnotify.Write:
					c.readPath(systemconfPath)
					c.announceReloaded()
				case fsnotify.Remove:
					c.resetToDefault()
					c.announceReloaded()
				}
			}
		case err, ok := <-c.watcher.Errors:
			if !ok {
				return
			}

			common.LogError("Failed to watch configuration file", err)
		}
	}
}

func (c *Config) resetToDefault() {
	c.reset()
	c.applyDefault()
}

func (c *Config) announceReloaded() {
	c.Backlight.announceReloaded(c.Backlight)
	c.Player.announceReloaded(c.Player)
	c.Volume.announceReloaded(c.Volume)
	c.SwayNodes.announceReloaded(c.SwayNodes)

	if c.notifier != nil {
		c.notifier.Notify("Swaypanion configuration reloaded")
	}
}

func (c *Config) applyDefault() {
	c.Backlight.applyDefault()
	c.Player.applyDefault()
	c.Volume.applyDefault()
	c.SwayNodes.applyDefault()

	c.CoreMessages = c.CoreMessages.applyDefault(Default.CoreMessages)
}

func (c *Config) reset() {
	c.Backlight = &Backlight{}
	c.Player = &Player{}
	c.Volume = &Volume{}
	c.SwayNodes = &SwayNodes{}

	c.CoreMessages = NotificationSectionMessage{}
}
