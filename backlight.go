package swaypanion

import (
	"bytes"
	"errors"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	sysBacklight           = "/sys/class/backlight"
	brightnessEntryFile    = "brightness"
	maxBrightnessEntryFile = "max_brightness"

	defaultBrightnessMinimum  = 1
	defaultBrightnessStepSize = 5
)

var ErrNoBacklightAvailable = errors.New("no backlight device available")

type BacklightConfig struct {
	Disable    bool   `yaml:"disable"`
	DeviceName string `yaml:"device_name"`
	Minimum    int    `yaml:"minimum"`
	StepSize   int    `yaml:"step_size"`
}

func (c BacklightConfig) withDefaults() BacklightConfig {
	if c.Minimum == 0 {
		c.Minimum = defaultBrightnessMinimum
	}

	if c.StepSize == 0 {
		c.StepSize = defaultBrightnessStepSize
	}

	return c
}

type Backlight struct {
	dataFilePath  string
	maximumRaw    int
	minimumRaw    int
	stepSizeRaw   float64
	onePercentRaw float64
}

func NewBacklight(conf BacklightConfig) (*Backlight, error) {
	if conf.Disable {
		return nil, nil
	}

	var deviceName string
	if conf.DeviceName == "" {
		entries, err := os.ReadDir(sysBacklight)
		if err != nil {
			return nil, err
		}

		if len(entries) == 0 {
			return nil, ErrNoBacklightAvailable
		}

		deviceName = entries[0].Name()
	} else {
		deviceName = conf.DeviceName
	}

	dataFilePath := filepath.Join(sysBacklight, deviceName, brightnessEntryFile)
	fileStat, err := os.Stat(dataFilePath)
	if err != nil {
		return nil, err
	}

	if fileStat.IsDir() {
		return nil, os.ErrInvalid
	}

	maxContent, err := os.ReadFile(filepath.Join(sysBacklight, deviceName, maxBrightnessEntryFile))
	if err != nil {
		return nil, err
	}

	maximumRaw, err := strconv.Atoi(strings.TrimSpace(string(maxContent)))
	if err != nil {
		return nil, err
	}

	onePercentRaw := float64(maximumRaw) / 100

	return &Backlight{
		dataFilePath:  dataFilePath,
		maximumRaw:    maximumRaw,
		minimumRaw:    int(math.Round(float64(conf.Minimum) * onePercentRaw)),
		stepSizeRaw:   float64(conf.StepSize) * onePercentRaw,
		onePercentRaw: onePercentRaw,
	}, nil
}

func (b *Backlight) currentRaw(roundedToStep bool) (float64, error) {
	currentContent, err := os.ReadFile(b.dataFilePath)
	if err != nil {
		return 0, err
	}

	current, err := strconv.ParseFloat(string(bytes.TrimSpace(currentContent)), 64)

	if roundedToStep {
		current = current / b.stepSizeRaw * b.stepSizeRaw
	}

	return current, err
}

func (b *Backlight) GetBrightness() (int, error) {
	value, err := b.currentRaw(false)
	if err != nil {
		return 0, err
	}

	return int(math.Round(value / b.onePercentRaw)), nil
}
func (b *Backlight) SetBrightness(percent int) error {
	percent = max(min(percent, 100), 0)

	value := int(math.Round(float64(percent) * b.onePercentRaw))

	return os.WriteFile(b.dataFilePath, []byte(strconv.Itoa(value)), 0x644)
}

func (b *Backlight) BrightnessUp() error {
	current, err := b.currentRaw(true)
	if err != nil {
		return err
	}

	target := min(
		int(math.Round(current+b.stepSizeRaw)),
		b.maximumRaw,
	)

	return os.WriteFile(b.dataFilePath, []byte(strconv.Itoa(target)), 0x644)
}

func (b *Backlight) BrightnessDown() error {
	current, err := b.currentRaw(true)
	if err != nil {
		return err
	}

	target := max(
		int(math.Round(current-b.stepSizeRaw)),
		b.minimumRaw,
	)

	return os.WriteFile(b.dataFilePath, []byte(strconv.Itoa(target)), 0x644)
}

func (s *Swaypanion) brightnessHandler(args []string, w io.Writer) error {
	if len(args) == 0 {
		b, err := s.backlight.GetBrightness()
		if err != nil {
			return err
		}

		return writeInt(w, b)
	}

	switch args[0] {
	case "up":
		if err := s.backlight.BrightnessUp(); err != nil {
			return err
		}
	case "down":
		if err := s.backlight.BrightnessDown(); err != nil {
			return err
		}
	case "set":
		if len(args) < 2 {
			return missingArgument("brightness value")
		}

		value, err := strconv.Atoi(args[1])
		if err != nil {
			return wrongIntArgument(args[1])
		}

		if err := s.backlight.SetBrightness(value); err != nil {
			return err
		}
	default:
		return s.unknownArgument(w, "brightness", args[0])
	}

	b, err := s.backlight.GetBrightness()
	if err != nil {
		return err
	}

	s.notification.NotifyLevel("brightness", b)

	return writeInt(w, b)
}

func (s *Swaypanion) registerBacklight() {
	if s.conf.Backlight.Disable {
		return
	}

	s.register(
		"brightness", s.brightnessHandler, "get or set the screen brightness",
		"", "show the screen brightness in percent",
		"up", "make the screen brighter",
		"down", "make the screen dimmer",
		"set X", "set the screen brightness to X percent",
	)
}
