package swaypanion

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	sysBacklight           = "/sys/class/backlight"
	brightnessEntryFile    = "brightness"
	maxBrightnessEntryFile = "max_brightness"

	defaultBrightnessMinimum      = 1
	defaultBrightnessStepSize     = 5
	defaultPollInterval           = 500 * time.Millisecond
	defaultBrightnessStatusFormat = `{"text": "<percent> %", "percentage": <percent>}`
)

var ErrNoBacklightAvailable = errors.New("no backlight device available")

type BacklightConfig struct {
	Disable      bool          `yaml:"disable"`
	DeviceName   string        `yaml:"device_name"`
	Minimum      uint          `yaml:"minimum"`
	StepSize     uint          `yaml:"step_size"`
	PollInterval time.Duration `yaml:"interval"`
	StatusFormat string        `yaml:"status_format"`
}

func (c BacklightConfig) withDefaults() BacklightConfig {
	if c.Minimum == 0 {
		c.Minimum = defaultBrightnessMinimum
	}

	if c.StepSize == 0 {
		c.StepSize = defaultBrightnessStepSize
	}

	if c.PollInterval == 0 {
		c.PollInterval = defaultPollInterval
	}

	if c.PollInterval > subscriptionDedupDelay {
		c.PollInterval = subscriptionDedupDelay
	}

	if c.StatusFormat == "" {
		c.StatusFormat = defaultBrightnessStatusFormat
	}

	return c
}

type Backlight struct {
	conf         BacklightConfig
	notification *Notification

	dataFilePath  string
	maximumRaw    int
	minimumRaw    int
	stepSizeRaw   float64
	onePercentRaw float64

	subscriptions *subscription[int]
	pollStop      chan bool
}

func NewBacklight(conf BacklightConfig, notification *Notification) (*Backlight, error) {
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

	b := &Backlight{
		conf:          conf,
		notification:  notification,
		dataFilePath:  dataFilePath,
		maximumRaw:    maximumRaw,
		minimumRaw:    int(math.Round(float64(conf.Minimum) * onePercentRaw)),
		stepSizeRaw:   float64(conf.StepSize) * onePercentRaw,
		onePercentRaw: onePercentRaw,
	}

	initial, err := b.get()
	if err != nil {
		return nil, err
	}

	b.subscriptions = newSubscription(initial)

	go b.poll()

	return b, nil
}

func (b *Backlight) Close() {
	if b.pollStop != nil {
		close(b.pollStop)
	}
}

func (b *Backlight) poll() {
	poller := time.NewTicker(b.conf.PollInterval)
	defer func() {
		poller.Stop()
		b.pollStop = nil
	}()
	b.pollStop = make(chan bool, 1)

	var knownBrightness string
	for {
		select {
		case <-poller.C:
			brightness, err := b.getRawData()
			if err != nil {
				slog.Error("Failed to get brightness", "error", err)
			}

			if brightness == knownBrightness {
				continue
			}

			// Avoid publishing before reading for the first time
			if knownBrightness != "" {
				raw, err := b.rawDataToInt(brightness)
				if err != nil {
					slog.Error("Failed to convert brightness", "error", err)
				}

				b.publish(raw)
			}

			knownBrightness = brightness
		case <-b.pollStop:
			return
		}
	}

}

func (b *Backlight) formatStatus(raw int) string {
	return strings.ReplaceAll(b.conf.StatusFormat, "<percent>", strconv.Itoa(b.rawToPercent(raw)))
}

func (b *Backlight) publish(raw int) {
	percent := b.rawToPercent(raw)

	if !b.subscriptions.Publish(percent) {
		return
	}

	if b.notification != nil {
		b.notification.NotifyLevel("brightness", percent)
	}
}

func (b *Backlight) rawToPercent(raw int) int {
	return int(math.Round(float64(raw) / b.onePercentRaw))
}

func (b *Backlight) percentToRaw(percent int) int {
	return int(math.Round(float64(percent) * b.onePercentRaw))
}

func (b *Backlight) roundToStep(raw int) int {
	return int(math.Round(float64(raw)/b.stepSizeRaw) * b.stepSizeRaw)
}

func (b *Backlight) rawDataToInt(data string) (int, error) {
	return strconv.Atoi(strings.TrimSpace(data))
}

func (b *Backlight) getRawData() (string, error) {
	data, err := os.ReadFile(b.dataFilePath)
	return string(data), err
}

func (b *Backlight) get() (int, error) {
	data, err := b.getRawData()
	if err != nil {
		return 0, err
	}

	return b.rawDataToInt(data)
}

func (b *Backlight) set(raw int) error {
	// Make sure value is between defined minimum and absolute maximum
	raw = max(min(raw, b.maximumRaw), b.minimumRaw)

	if err := os.WriteFile(b.dataFilePath, []byte(strconv.Itoa(raw)), 0x644); err != nil {
		return err
	}

	b.publish(raw)

	return nil
}

func (b *Backlight) GetBrightness() (int, error) {
	value, err := b.get()
	if err != nil {
		return 0, err
	}

	return b.rawToPercent(value), nil
}

func (b *Backlight) SetBrightness(percent int) error {
	return b.set(b.percentToRaw(percent))
}

func (b *Backlight) BrightnessUp() error {
	current, err := b.get()
	if err != nil {
		return err
	}

	return b.set(b.roundToStep(current) + int(b.stepSizeRaw))
}

func (b *Backlight) BrightnessDown() error {
	current, err := b.get()
	if err != nil {
		return err
	}

	return b.set(b.roundToStep(current) - int(b.stepSizeRaw))
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
		return s.backlight.BrightnessUp()
	case "down":
		return s.backlight.BrightnessDown()
	case "set":
		if len(args) < 2 {
			return missingArgument("brightness value")
		}

		value, err := strconv.Atoi(args[1])
		if err != nil {
			return wrongIntArgument(args[1])
		}

		return s.backlight.SetBrightness(value)
	case "subscribe":
		status, id := s.backlight.subscriptions.Subscribe()

		for b := range status {
			if err := writeString(w, s.backlight.formatStatus(b)); err != nil {
				s.backlight.subscriptions.Unsubscribe(id)

				if strings.Contains(err.Error(), "broken pipe") {
					return nil
				}

				return fmt.Errorf("failed to send brightness to client: %w", err)

			}
		}

		return nil
	default:
		return s.unknownArgument(w, "brightness", args[0])
	}
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
		"subscribe", "receive the brightness as soon as it changes",
	)
}
