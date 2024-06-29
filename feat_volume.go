package swaypanion

import (
	"fmt"
	"io"
	"log/slog"
	"math"
	"net"
	"strconv"
	"strings"

	"github.com/jfreymuth/pulse/proto"
)

const (
	volumeMaximumRaw      = 65536
	volumeOnePercentRaw   = float64(volumeMaximumRaw) / 100
	defaultVolumeSinkName = "@DEFAULT_SINK@"
	defaultVolumeStepSize = 5
)

type VolumeConfig struct {
	Disable  bool   `yaml:"disable"`
	Server   string `yaml:"server"`
	SinkName string `yaml:"sink_name"`
	StepSize int    `yaml:"step_size"`
}

func (c VolumeConfig) withDefaults() VolumeConfig {
	if c.SinkName == "" {
		c.SinkName = defaultVolumeSinkName
	}

	if c.StepSize == 0 {
		c.StepSize = defaultVolumeStepSize
	}

	return c
}

type Volume struct {
	conf         VolumeConfig
	notification *Notification

	stepSizeRaw float64

	paClient  *proto.Client
	paConn    net.Conn
	sinkIndex uint32

	subscriptions *subscription[int]
}

func NewVolume(conf VolumeConfig, notification *Notification) (*Volume, error) {
	client, conn, err := proto.Connect(conf.Server)
	if err != nil {
		return nil, err
	}

	v := &Volume{
		conf:         conf,
		notification: notification,
		stepSizeRaw:  float64(conf.StepSize) * volumeOnePercentRaw,
		paClient:     client,
		paConn:       conn,
	}

	if err := v.paClient.Request(&proto.SetClientName{Props: proto.PropList{}}, nil); err != nil {
		return nil, err
	}

	lookupSink := proto.LookupSinkReply{}
	if err := v.paClient.Request(&proto.LookupSink{SinkName: v.conf.SinkName}, &lookupSink); err != nil {
		return nil, err
	}

	v.sinkIndex = lookupSink.SinkIndex
	v.paClient.Callback = v.paCallback

	initial, err := v.GetVolume()
	if err != nil {
		return nil, fmt.Errorf("failed to read volume: %w", err)
	}

	v.subscriptions = newSubscription(initial)

	if err := v.paClient.Request(&proto.Subscribe{Mask: proto.SubscriptionMaskSink}, nil); err != nil {
		return nil, err
	}

	return v, nil
}

func (v *Volume) Close() error {
	return v.paConn.Close()
}

func (v *Volume) paCallback(val any) {
	evt, ok := val.(*proto.SubscribeEvent)
	if !ok {
		return
	}

	if evt.Event.GetType() == proto.EventChange && evt.Event.GetFacility() == proto.EventSink && evt.Index == v.sinkIndex {
		go v.announceVolume()
	}
}

func (v *Volume) currentRaw(roundedToStep bool) (float64, bool, error) {
	repl := proto.GetSinkInfoReply{}

	if err := v.paClient.Request(
		&proto.GetSinkInfo{SinkIndex: v.sinkIndex},
		&repl,
	); err != nil {
		return 0, false, err
	}

	var currentTotal uint32

	// Take the volume as the average across all channels.
	for _, ch := range repl.ChannelVolumes {
		currentTotal += ch
	}

	current := float64(currentTotal) / float64(len(repl.ChannelVolumes))

	if roundedToStep {
		current = math.Round(current/v.stepSizeRaw) * v.stepSizeRaw
	}

	return current, repl.Mute, nil
}

func (v *Volume) GetVolume() (volume int, err error) {
	value, mute, err := v.currentRaw(false)
	if err != nil {
		return 0, err
	}

	if mute {
		return -1, nil
	}

	return int(math.Round(value / volumeOnePercentRaw)), nil
}

func (v *Volume) set(value int) error {
	return v.paClient.Request(&proto.SetSinkVolume{
		SinkIndex:      v.sinkIndex,
		ChannelVolumes: proto.ChannelVolumes{uint32(value)},
	}, nil)
}

func (v *Volume) SetVolume(percent int) error {
	percent = max(min(percent, 100), 0)

	target := min(
		int(math.Round(float64(percent)*volumeOnePercentRaw)),
		volumeMaximumRaw,
	)

	return v.set(target)
}

func (v *Volume) VolumeUp() error {
	current, _, err := v.currentRaw(true)
	if err != nil {
		return err
	}

	target := min(
		int(math.Round(current+v.stepSizeRaw)),
		volumeMaximumRaw,
	)

	return v.set(target)
}

func (v *Volume) VolumeDown() error {
	current, _, err := v.currentRaw(true)
	if err != nil {
		return err
	}

	target := max(
		int(math.Round(current-v.stepSizeRaw)),
		0,
	)

	return v.set(target)
}

func (v *Volume) ToggleMute() error {
	volume, err := v.GetVolume()
	if err != nil {
		return err
	}

	return v.paClient.Request(&proto.SetSinkMute{
		SinkIndex: v.sinkIndex,
		Mute:      volume != -1,
	}, nil)
}

func (v *Volume) announceVolume() {
	volume, err := v.GetVolume()
	if err != nil {
		slog.Error("Failed to read volume", "error", err)
		return
	}

	if !v.subscriptions.Publish(volume) {
		return
	}

	if v.notification != nil {
		switch {
		case volume == -1:
			v.notification.NotifyLevel("volume_muted", -1)
		case volume < 50:
			v.notification.NotifyLevel("volume_low", volume)
		default:
			v.notification.NotifyLevel("volume_high", volume)
		}
	}
}

func (s *Swaypanion) volumeHandler(args []string, w io.Writer) error {
	if len(args) == 0 {
		volume, err := s.volume.GetVolume()
		if err != nil {
			return err
		}

		if volume == -1 {
			return writeString(w, "muted")
		}

		return writeInt(w, volume)
	}

	switch args[0] {
	case "up":
		return s.volume.VolumeUp()
	case "down":
		return s.volume.VolumeDown()
	case "mute":
		return s.volume.ToggleMute()
	case "set":
		if len(args) < 2 {
			return missingArgument("volume value")
		}

		value, err := strconv.Atoi(args[1])
		if err != nil {
			return wrongIntArgument(args[1])
		}

		return s.volume.SetVolume(value)
	case "subscribe":
		volume, id := s.volume.subscriptions.Subscribe()

		for vol := range volume {
			var err error
			if vol == -1 {
				err = writeString(w, "muted")
			} else {
				err = writeInt(w, vol)
			}

			if err != nil {
				s.volume.subscriptions.Unsubscribe(id)

				if strings.Contains(err.Error(), "broken pipe") {
					return nil
				}

				return fmt.Errorf("failed to send volume to client: %w", err)

			}
		}

		return nil
	default:
		return s.unknownArgument(w, "volume", args[0])
	}
}

func (s *Swaypanion) registerVolume() {
	if s.conf.Volume.Disable {
		return
	}

	s.register(
		"volume", s.volumeHandler, "get or set the volume",
		"", "show the volume in percent",
		"up", "make the volume higher",
		"down", "make the volume lower",
		"mute", "toggle the mute status",
		"set X", "set the volume to X percent",
		"subscribe", "receive the volume as soon as it changes",
	)
}
