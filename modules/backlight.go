package modules

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/willoma/swaypanion/common"
	"github.com/willoma/swaypanion/config"
	"github.com/willoma/swaypanion/notification"
)

type Backlight struct {
	subscriptions *common.Pubsub[common.Int]
	notifier      *notification.PercentNotifier

	stop func()

	mu            sync.Mutex
	working       bool
	dataFilePath  string
	minimumRaw    int
	maximumRaw    int
	onePercentRaw float64
	stepRaw       float64
}

func NewBacklight(conf *config.Backlight, notif *notification.Notification) *Backlight {
	b := &Backlight{
		subscriptions: common.NewPubsub[common.Int](),
		notifier:      notif.PercentNotifier(),
	}

	b.reloadConfig(conf)
	b.stop = conf.ListenReload(b.reloadConfig)

	b.subscriptions.Subscribe(b.notifier, false, b.notifier.Notify)

	return b
}

func (b *Backlight) Stop() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.stop != nil {
		b.stop()
		b.stop = nil
	}

	b.subscriptions.Unsubscribe(b.notifier)
}

func (b *Backlight) reloadConfig(conf *config.Backlight) {
	b.mu.Lock()
	defer b.mu.Unlock()

	var deviceName string

	if conf.DeviceName == "" {
		entries, err := os.ReadDir("/sys/class/backlight")
		if err != nil {
			common.LogError("Failed to search backlight device", err)
			return
		}

		if len(entries) == 0 {
			common.LogError("No backlight device available", nil)
			return
		}

		deviceName = entries[0].Name()
	} else {
		deviceName = conf.DeviceName
	}

	dataFilePath := filepath.Join("/sys/class/backlight", deviceName, "brightness")

	if dataFilePath == b.dataFilePath {
		return
	}

	b.working = false

	fileStat, err := os.Stat(dataFilePath)
	if err != nil {
		common.LogError("Failed to check brightness pseudofile", err)
		return
	}

	if fileStat.IsDir() {
		common.LogError("Brightness pseudofile is a directory", nil)
		return
	}

	maxContent, err := os.ReadFile(filepath.Join("/sys/class/backlight", deviceName, "max_brightness"))
	if err != nil {
		common.LogError("Failed to read maximum brightness pseudofile", err)
		return
	}

	b.maximumRaw, err = strconv.Atoi(strings.TrimSpace(string(maxContent)))
	if err != nil {
		common.LogError("Failed to read maximum brightness", err)
		return
	}

	b.dataFilePath = dataFilePath
	b.onePercentRaw = onePercent(b.maximumRaw)
	b.minimumRaw = round[int](conf.MinimumPercent * b.onePercentRaw)
	b.stepRaw = conf.StepSizePercent * b.onePercentRaw

	b.subscriptions.Reconfigure(common.Config[common.Int]{
		PollInterval: conf.PollInterval,
		PollFn: common.IntPoller(func() (value int, disabled bool, ok bool) {
			value, ok = b.get()
			return
		}),
	})

	b.notifier.Reconfigure(conf.Notification)

	b.working = true
}

func (b *Backlight) unsafeGetRaw() (int, bool) {
	if !b.working {
		return 0, false
	}

	data, err := os.ReadFile(b.dataFilePath)
	if err != nil {
		common.LogError("Failed to get brightness", err)
		return 0, false
	}

	value, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		common.LogError("Failed to convert raw data to integer", err)
		return 0, false
	}

	return value, true
}

func (b *Backlight) unsafeSetRawAndGetPercent(raw int) (percent int, ok bool) {
	if !b.working {
		return 0, false
	}

	raw = limit(raw, b.minimumRaw, b.maximumRaw)

	if err := os.WriteFile(b.dataFilePath, []byte(strconv.Itoa(raw)), 0x644); err != nil {
		common.LogError("Failed to set brightness", err)
		return 0, false
	}

	percent = toPercent(raw, b.maximumRaw)

	b.subscriptions.Publish(common.Int{Value: percent})

	return percent, true
}

func (b *Backlight) get() (int, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	raw, ok := b.unsafeGetRaw()
	return toPercent(raw, b.maximumRaw), ok
}

func (b *Backlight) set(percent int) (int, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	raw := round[int](float64(percent) * b.onePercentRaw)

	return b.unsafeSetRawAndGetPercent(raw)
}

func (b *Backlight) up() (int, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	raw, ok := b.unsafeGetRaw()
	if !ok {
		return 0, false
	}

	raw = int(roundStep(float64(raw)+b.stepRaw, b.stepRaw))

	return b.unsafeSetRawAndGetPercent(raw)
}

func (b *Backlight) down() (int, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	raw, ok := b.unsafeGetRaw()
	if !ok {
		return 0, false
	}

	raw = int(roundStep(float64(raw)-b.stepRaw, b.stepRaw))

	return b.unsafeSetRawAndGetPercent(raw)
}
