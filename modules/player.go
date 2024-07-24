package modules

import (
	"errors"
	"strings"
	"sync"

	"github.com/godbus/dbus/v5"
	"github.com/willoma/swaypanion/common"
	"github.com/willoma/swaypanion/config"
	"github.com/willoma/swaypanion/sway"
)

var ErrNoPlayer = errors.New("no player found")

type playerData struct {
	Status  string
	Artists []string
	Album   string
	Title   string
}

func (p playerData) Equal(o playerData) bool {
	if p.Status != o.Status {
		return false
	}

	if len(p.Artists) != len(o.Artists) {
		return false
	}

	for i := range p.Artists {
		if p.Artists[i] != o.Artists[i] {
			return false
		}
	}

	if p.Album != o.Album {
		return false
	}

	if p.Title != o.Title {
		return false
	}

	return true
}

func (p playerData) complement() []string {
	complement := []string{}

	if len(p.Artists) > 0 && p.Artists[0] != "" {
		complement = append(complement, "Artist: "+p.Artists[0])
	}

	if len(p.Artists) > 1 {
		complement = append(complement, "Artists: "+strings.Join(p.Artists, ", "))
	}

	if p.Album != "" {
		complement = append(complement, "Album: "+p.Album)
	}

	if p.Title != "" {
		complement = append(complement, "Title: "+p.Title)
	}

	return complement
}

type Player struct {
	sway          *sway.Client
	subscriptions *common.Pubsub[playerData]

	stop func()

	mu          sync.Mutex
	working     bool
	dbus        *dbus.Conn
	start       config.Command
	show        config.Command
	signalCh    chan *dbus.Signal
	playerName  string
	currentData playerData
}

func NewPlayer(conf *config.Player, swayClient *sway.Client) *Player {
	p := &Player{
		sway:          swayClient,
		subscriptions: common.NewPubsub[playerData](),
	}

	p.reloadConfig(conf)
	p.stop = conf.ListenReload(p.reloadConfig)

	// Load current values
	if data, ok := p.get(); ok {
		p.subscriptions.Publish(data)
	}

	return p
}

func (p *Player) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.stop != nil {
		p.stop()
		p.stop = nil
	}

	p.unsafeStopDbusSignal()
}

func (p *Player) reloadConfig(conf *config.Player) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.working = false

	p.unsafeStopDbusSignal()

	dbusConn, err := dbus.SessionBus()
	if err != nil {
		common.LogError("Failed to connect to DBus", err)
		return
	}

	p.dbus = dbusConn

	if err := p.dbus.AddMatchSignal(
		dbus.WithMatchObjectPath("/org/mpris/MediaPlayer2"),
		dbus.WithMatchInterface("org.freedesktop.DBus.Properties"),
	); err != nil {
		common.LogError("Failed to add signal", err)
	}

	p.playerName = conf.PlayerName
	p.start = conf.Start
	p.show = conf.Show

	p.working = true

	go p.listenDBusSignal()
}

func (p *Player) unsafeStopDbusSignal() {
	if p.signalCh != nil {
		p.dbus.RemoveSignal(p.signalCh)
		p.signalCh = nil

		p.dbus.RemoveMatchSignal(
			dbus.WithMatchObjectPath("/org/mpris/MediaPlayer2"),
			dbus.WithMatchInterface("org.freedesktop.DBus.Properties"),
		)

	}
}

func (p *Player) listenDBusSignal() {
	p.signalCh = make(chan *dbus.Signal, 10)
	p.dbus.Signal(p.signalCh)

	for val := range p.signalCh {
		// https://dbus.freedesktop.org/doc/dbus-specification.html#standard-interfaces-properties
		changedProperties, ok := val.Body[1].(map[string]dbus.Variant)
		if !ok {
			continue
		}

		p.mu.Lock()

		if status, ok := changedProperties["PlaybackStatus"]; ok {
			p.unsafeStoreStatus(status.Value())
		}

		if metadata, ok := changedProperties["Metadata"]; ok {
			p.unsafeStoreMetadata(metadata.Value())
		}

		p.subscriptions.Publish(p.currentData)

		p.mu.Unlock()
	}
}

func (p *Player) unsafeCall(method string, args ...any) (any, error) {
	call := p.dbus.Object(
		"org.mpris.MediaPlayer2."+p.playerName,
		"/org/mpris/MediaPlayer2",
	).Call(method, 0, args...)

	if call.Err != nil {
		if call.Err.Error() == "The name org.mpris.MediaPlayer2."+
			p.playerName+
			" was not provided by any .service files" {
			return nil, ErrNoPlayer
		}

		return nil, call.Err
	}

	if len(call.Body) > 0 {
		if response, ok := call.Body[0].(dbus.Variant); ok {
			return response.Value(), nil
		}
	}

	return nil, nil
}

func (p *Player) unsafeStoreMetadata(data any) {
	p.currentData.Album = ""
	p.currentData.Artists = nil
	p.currentData.Title = ""

	array, ok := data.(map[string]dbus.Variant)
	if !ok {
		return
	}

	for k, v := range array {
		switch k {
		case "xesam:album":
			album, _ := v.Value().(string)
			p.currentData.Album = album
		case "xesam:artist":
			artists, _ := v.Value().([]string)
			p.currentData.Artists = artists
		case "xesam:title":
			title, _ := v.Value().(string)
			p.currentData.Title = title
		}
	}
}

func (p *Player) unsafeStoreStatus(data any) {
	statusValue, _ := data.(string)
	p.currentData.Status = statusValue
}

func (p *Player) get() (playerData, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	status, err := p.unsafeCall("org.freedesktop.DBus.Properties.Get", "org.mpris.MediaPlayer2.Player", "PlaybackStatus")
	if err != nil {
		if errors.Is(err, ErrNoPlayer) {
			p.currentData.Status = err.Error()
			p.currentData.Album = ""
			p.currentData.Artists = nil
			p.currentData.Title = ""

			return p.currentData, true
		}

		common.LogError("Failed to get player status", err)
		return playerData{}, false
	}

	metadata, err := p.unsafeCall("org.freedesktop.DBus.Properties.Get", "org.mpris.MediaPlayer2.Player", "Metadata")
	if err != nil {
		common.LogError("Failed to get player metadata", err)
		return playerData{}, false
	}

	p.unsafeStoreMetadata(metadata)
	p.unsafeStoreStatus(status)

	return p.currentData, true
}

func (p *Player) playpause() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	_, err := p.unsafeCall("org.mpris.MediaPlayer2.Player.PlayPause")
	return err
}

func (p *Player) previous() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	_, err := p.unsafeCall("org.mpris.MediaPlayer2.Player.Previous")
	return err
}

func (p *Player) next() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	_, err := p.unsafeCall("org.mpris.MediaPlayer2.Player.Next")
	return err
}

func (p *Player) startOrShow() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.currentData.Status == ErrNoPlayer.Error() {
		return p.start.Run(p.sway)
	}

	return p.show.Run(p.sway)
}
