package swaypanion

import (
	"io"
	"math"
	"net"
	"strconv"

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
	conf VolumeConfig

	stepSizeRaw float64

	paClient *proto.Client
	paConn   net.Conn
}

func NewVolume(conf VolumeConfig) (*Volume, error) {
	client, conn, err := proto.Connect(conf.Server)
	if err != nil {
		return nil, err
	}

	props := proto.PropList{}
	err = client.Request(&proto.SetClientName{Props: props}, nil)
	if err != nil {
		panic(err)
	}

	return &Volume{
		conf:        conf,
		stepSizeRaw: float64(conf.StepSize) * volumeOnePercentRaw,
		paClient:    client,
		paConn:      conn,
	}, nil
}

func (v *Volume) currentRaw(roundedToStep bool) (float64, bool, error) {
	repl := proto.GetSinkInfoReply{}

	if err := v.paClient.Request(
		&proto.GetSinkInfo{SinkIndex: proto.Undefined, SinkName: v.conf.SinkName},
		&repl,
	); err != nil {
		return 0, false, err
	}

	var current float64

	// Take the volume as the average across all channels.
	for _, ch := range repl.ChannelVolumes {
		current += float64(ch)
	}

	current = current / float64(len(repl.ChannelVolumes))

	if roundedToStep {
		current = current / v.stepSizeRaw * v.stepSizeRaw
	}

	return current, repl.Mute, nil
}

func (v *Volume) GetVolume() (volume int, muted bool, err error) {
	value, mute, err := v.currentRaw(false)
	if err != nil {
		return 0, false, err
	}

	return int(math.Round(value / volumeOnePercentRaw)), mute, nil
}

func (v *Volume) set(value int) error {
	return v.paClient.Request(&proto.SetSinkVolume{
		SinkIndex:      proto.Undefined,
		SinkName:       v.conf.SinkName,
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
	_, muted, err := v.GetVolume()
	if err != nil {
		return err
	}

	return v.paClient.Request(&proto.SetSinkMute{
		SinkIndex: proto.Undefined,
		SinkName:  v.conf.SinkName,
		Mute:      !muted,
	}, nil)
}

func (v *Volume) Close() error {
	return v.paConn.Close()
}

func (s *Swaypanion) volumeHandler(args []string, w io.Writer) error {
	if len(args) == 0 {
		volume, muted, err := s.volume.GetVolume()
		if err != nil {
			return err
		}

		if muted {
			_, err := w.Write([]byte("muted"))
			return err
		}

		return writeInt(w, volume)
	}

	switch args[0] {
	case "up":
		if err := s.volume.VolumeUp(); err != nil {
			return err
		}
	case "down":
		if err := s.volume.VolumeDown(); err != nil {
			return err
		}
	case "mute":
		if err := s.volume.ToggleMute(); err != nil {
			return err
		}
	case "set":
		if len(args) < 2 {
			return missingArgument("volume value")
		}

		value, err := strconv.Atoi(args[1])
		if err != nil {
			return wrongIntArgument(args[1])
		}

		if err := s.volume.SetVolume(value); err != nil {
			return err
		}
	default:
		return s.unknownArgument(w, "volume", args[0])
	}

	volume, muted, err := s.volume.GetVolume()
	if err != nil {
		return err
	}

	if muted {
		s.notification.NotifyLevel("volume_muted", -1)
		_, err := w.Write([]byte("muted"))
		return err
	}

	if volume < 50 {
		s.notification.NotifyLevel("volume_low", volume)
	} else {
		s.notification.NotifyLevel("volume_high", volume)
	}

	return writeInt(w, volume)
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
	)
}
