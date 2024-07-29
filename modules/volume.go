package modules

import (
	"net"
	"sync"

	"github.com/jfreymuth/pulse/proto"
	"github.com/willoma/swaypanion/common"
	"github.com/willoma/swaypanion/config"
	"github.com/willoma/swaypanion/notification"
)

const (
	paVolumeMaximumRaw    = 65536
	paVolumeOnePercentRaw = float64(paVolumeMaximumRaw) / 100
)

type Volume struct {
	subscriptions *common.Pubsub[common.Int]
	notifier      *notification.PercentNotifier

	stop func()

	mu        sync.Mutex
	working   bool
	paClient  *proto.Client
	paConn    net.Conn
	sinkIndex uint32
	stepRaw   float64
}

func NewVolume(conf *config.Volume, notif *notification.Notification) *Volume {
	v := &Volume{
		subscriptions: common.NewPubsub[common.Int](),
		notifier:      notif.PercentNotifier(),
	}

	v.reloadConfig(conf)
	v.stop = conf.ListenReload(v.reloadConfig)

	v.subscriptions.Subscribe(v.notifier, false, v.notifier.Notify)

	return v
}

func (v *Volume) Stop() {
	v.mu.Lock()
	defer v.mu.Unlock()

	if v.stop != nil {
		v.stop()
		v.stop = nil
	}

	v.subscriptions.Unsubscribe(v.notifier)
}

func (v *Volume) reloadConfig(conf *config.Volume) {
	v.mu.Lock()
	defer v.mu.Unlock()

	if v.paConn != nil {
		if err := v.paConn.Close(); err != nil {
			common.LogError("Failed to close previous pulseaudio connection", err)
		}
	}

	v.stepRaw = float64(conf.StepSize) * paVolumeOnePercentRaw

	v.working = false

	client, conn, err := proto.Connect(conf.Server)
	if err != nil {
		common.LogError("Failed to connect to pulseaudio", err)
		return
	}

	if err := client.Request(&proto.SetClientName{Props: proto.PropList{}}, nil); err != nil {
		common.LogError("Failed to initialize pulseaudio connection", err)
		return
	}

	lookupSink := proto.LookupSinkReply{}
	if err := client.Request(&proto.LookupSink{SinkName: conf.SinkName}, &lookupSink); err != nil {
		common.LogError("Sink not found in pulseaudio server (sink "+conf.SinkName+")", err)
		return
	}

	client.Callback = v.paCallback

	if err := client.Request(&proto.Subscribe{Mask: proto.SubscriptionMaskSink}, nil); err != nil {
		common.LogError("Failed to subscribe to pulseaudio events", err)
		return
	}

	v.sinkIndex = lookupSink.SinkIndex
	v.paClient = client
	v.paConn = conn

	v.notifier.Reconfigure(conf.Notification)

	v.working = true

	if volume, mute, ok := v.unsafeGet(); ok {
		v.subscriptions.Publish(common.Int{Disabled: mute, Value: volume})
	}
}

func (v *Volume) paCallback(val any) {
	evt, ok := val.(*proto.SubscribeEvent)
	if !ok {
		return
	}

	if evt.Event.GetType() == proto.EventChange && evt.Event.GetFacility() == proto.EventSink && evt.Index == v.sinkIndex {
		go func() {
			if volume, mute, ok := v.get(); ok {
				v.subscriptions.Publish(common.Int{Disabled: mute, Value: volume})
			}
		}()
	}
}

func (v *Volume) unsafeGetRaw() (volume float64, mute bool, ok bool) {
	if !v.working {
		return 0, false, false
	}

	repl := proto.GetSinkInfoReply{}

	if err := v.paClient.Request(
		&proto.GetSinkInfo{SinkIndex: v.sinkIndex},
		&repl,
	); err != nil {
		common.LogError("Failed to get volume", err)
		return 0, false, false
	}

	var currentTotal uint32

	// Take the volume as the average across all channels.
	for _, ch := range repl.ChannelVolumes {
		currentTotal += ch
	}

	current := float64(currentTotal) / float64(len(repl.ChannelVolumes))

	return current, repl.Mute, true
}

func (v *Volume) unsafeSetRawAndGetPercent(raw float64) (percent int, ok bool) {
	if !v.working {
		return 0, false
	}

	raw = limit(raw, 0, paVolumeMaximumRaw)

	if err := v.paClient.Request(&proto.SetSinkVolume{
		SinkIndex:      v.sinkIndex,
		ChannelVolumes: proto.ChannelVolumes{round[uint32](raw)},
	}, nil); err != nil {
		common.LogError("Failed to set volume", err)
		return 0, false
	}

	percent = toPercent(raw, paVolumeMaximumRaw)

	v.subscriptions.Publish(common.Int{Disabled: false, Value: percent})

	return percent, true
}

func (v *Volume) get() (vol int, mute, ok bool) {
	v.mu.Lock()
	defer v.mu.Unlock()

	return v.unsafeGet()
}

func (v *Volume) unsafeGet() (vol int, mute, ok bool) {
	raw, mute, ok := v.unsafeGetRaw()

	return toPercent(raw, paVolumeMaximumRaw), mute, ok
}

func (v *Volume) set(percent int) (int, bool) {
	v.mu.Lock()
	defer v.mu.Unlock()

	raw := float64(percent) * paVolumeOnePercentRaw

	return v.unsafeSetRawAndGetPercent(raw)
}

func (v *Volume) up() (int, bool) {
	v.mu.Lock()
	defer v.mu.Unlock()

	raw, _, ok := v.unsafeGetRaw()
	if !ok {
		return 0, false
	}

	raw = roundStep(raw+v.stepRaw, v.stepRaw)

	return v.unsafeSetRawAndGetPercent(raw)
}

func (v *Volume) down() (int, bool) {
	v.mu.Lock()
	defer v.mu.Unlock()

	raw, _, ok := v.unsafeGetRaw()
	if !ok {
		return 0, false
	}

	raw = roundStep(raw-v.stepRaw, v.stepRaw)

	return v.unsafeSetRawAndGetPercent(raw)
}

func (v *Volume) mute() (volume int, mute, ok bool) {
	v.mu.Lock()
	defer v.mu.Unlock()

	volume, mute, ok = v.unsafeGet()
	if !ok {
		return 0, false, false
	}

	if err := v.paClient.Request(&proto.SetSinkMute{
		SinkIndex: v.sinkIndex,
		Mute:      !mute,
	}, nil); err != nil {
		common.LogError("Failed to change mute status", err)
		return volume, mute, false
	}

	v.subscriptions.Publish(common.Int{Disabled: !mute, Value: volume})

	return volume, !mute, true
}
