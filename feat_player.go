package swaypanion

import (
	"errors"
	"fmt"
	"html"
	"io"
	"strings"
	"sync"

	"github.com/godbus/dbus/v5"
)

const (
	playerMessagePrefix = "org.mpris.MediaPlayer2.Player"
	playerObjectPath    = "/org/mpris/MediaPlayer2"

	defaultPlayerName           = "spotify"
	defaultPlayerCmdType        = CommandTypeSway
	defaultPlayerLaunchCommand  = "exec /usr/bin/spotify"
	defaultPlayerShowCommand    = "[instance=spotify] move workspace current, focus"
	defaultPlayerDiscLabel      = " "
	defaultPlayerNoPlayerStatus = `{"text": " "}`
	defaultPlayerStatusFormat   = `{"alt": "<status>", "text": "<title>", "tooltip": "<artists> - <title>\r <album>"}`
	defaultPlayerEscapeHTML     = true
)

var ErrNoPlayerFound = errors.New("no player found")

type PlayerConfig struct {
	Disable        bool        `yaml:"disable"`
	Name           string      `yaml:"name"`
	CmdType        CommandType `yaml:"command_type"`
	LaunchCommand  string      `yaml:"launch_command"`
	ShowCommand    string      `yaml:"show_command"`
	StatusFormat   string      `yaml:"status_format"`
	EscapeHTML     *bool       `yaml:"escape_html"`
	NoPlayerStatus string      `yaml:"no_player_status"`
}

func (c PlayerConfig) withDefaults() PlayerConfig {
	if c.Name == "" {
		c.Name = defaultPlayerName
	}

	c.CmdType = c.CmdType.defaultIfInvalid()

	if c.LaunchCommand == "" {
		c.LaunchCommand = defaultPlayerLaunchCommand
	}

	if c.ShowCommand == "" {
		c.ShowCommand = defaultPlayerShowCommand
	}

	if c.StatusFormat == "" {
		c.StatusFormat = defaultPlayerStatusFormat
	}

	if c.EscapeHTML == nil {
		v := defaultPlayerEscapeHTML
		c.EscapeHTML = &v
	}

	if c.NoPlayerStatus == "" {
		c.NoPlayerStatus = defaultPlayerNoPlayerStatus
	}

	return c
}

type playerMetadata struct {
	album   string
	artists []string
	title   string
}

type Player struct {
	conf PlayerConfig

	sway *SwayClient
	dbus *dbus.Conn

	signalCh chan *dbus.Signal

	currentMu       sync.Mutex
	currentMetadata playerMetadata
	currentStatus   string

	subscriptions *subscription[string]
}

func NewPlayer(conf PlayerConfig, sway *SwayClient) (*Player, error) {
	p := &Player{
		conf:     conf,
		sway:     sway,
		signalCh: make(chan *dbus.Signal, 10),
	}

	var err error
	p.dbus, err = dbus.SessionBus()
	if err != nil {
		return nil, err
	}

	if err := p.dbus.AddMatchSignal(
		dbus.WithMatchObjectPath(playerObjectPath),
		dbus.WithMatchInterface("org.freedesktop.DBus.Properties"),
	); err != nil {
		return nil, err
	}

	initial, err := p.FormattedStatus()
	if err != nil {
		return nil, err
	}

	p.subscriptions = newSubscription(initial)

	go p.dbusSignal()

	return p, nil
}

func (p *Player) Close() {
	p.dbus.RemoveSignal(p.signalCh)
	p.dbus.RemoveMatchSignal(
		dbus.WithMatchObjectPath(playerObjectPath),
		dbus.WithMatchInterface("org.freedesktop.DBus.Properties"),
	)
}

func (p *Player) dbusSignal() {
	p.dbus.Signal(p.signalCh)
	for val := range p.signalCh {
		// https://dbus.freedesktop.org/doc/dbus-specification.html#standard-interfaces-properties
		changedProperties, ok := val.Body[1].(map[string]dbus.Variant)
		if !ok {
			continue
		}

		p.currentMu.Lock()

		if status, ok := changedProperties["PlaybackStatus"]; ok {
			if statusValue, ok := status.Value().(string); ok {
				p.currentStatus = statusValue
			}
		}

		if metadata, ok := changedProperties["Metadata"]; ok {
			p.currentMetadata = dbusExtractPlayerMetadata(metadata.Value())
		}

		p.subscriptions.Publish(p.unsafeFormatStatus())

		p.currentMu.Unlock()
	}
}

func (p *Player) Launch() error {
	return execCmd(p.sway, p.conf.CmdType, p.conf.LaunchCommand)
}

func (p *Player) Show() error {
	return execCmd(p.sway, p.conf.CmdType, p.conf.ShowCommand)
}

func (p *Player) Next() error {
	_, err := p.call("org.mpris.MediaPlayer2.Player.Next")
	return err
}

func (p *Player) PlayPause() error {
	_, err := p.call("org.mpris.MediaPlayer2.Player.PlayPause")
	return err
}

func (p *Player) Previous() error {
	_, err := p.call("org.mpris.MediaPlayer2.Player.Previous")
	return err
}

func (p *Player) unsafeUpdateMetadata() error {
	response, err := p.call("org.freedesktop.DBus.Properties.Get", playerMessagePrefix, "Metadata")
	if err != nil {
		return err
	}

	p.currentMetadata = dbusExtractPlayerMetadata(response)

	return nil
}

func (p *Player) unsafeUpdatePlaybackStatus() error {
	response, err := p.call("org.freedesktop.DBus.Properties.Get", playerMessagePrefix, "PlaybackStatus")
	if err != nil {
		return err
	}

	if v, ok := response.(string); ok {
		p.currentStatus = v
	}

	return nil
}

func (p *Player) unsafeFormatStatus() string {
	var artist string
	if len(p.currentMetadata.artists) > 0 {
		artist = p.currentMetadata.artists[0]
	}
	artists := strings.Join(p.currentMetadata.artists, ", ")

	if *p.conf.EscapeHTML {
		return strings.NewReplacer(
			"<status>", p.currentStatus,
			"<title>", html.EscapeString(p.currentMetadata.title),
			"<artist>", html.EscapeString(artist),
			"<artists>", html.EscapeString(artists),
			"<album>", html.EscapeString(p.currentMetadata.album),
			"\\n", "\n",
			"\\r", "\r",
		).Replace(p.conf.StatusFormat)
	} else {
		return strings.NewReplacer(
			"<status>", p.currentStatus,
			"<title>", p.currentMetadata.title,
			"<artist>", artist,
			"<artists>", artists,
			"<album>", p.currentMetadata.album,
			"\\n", "\n",
			"\\r", "\r",
		).Replace(p.conf.StatusFormat)
	}
}

func (p *Player) FormattedStatus() (string, error) {
	p.currentMu.Lock()
	defer p.currentMu.Unlock()

	if err := p.unsafeUpdatePlaybackStatus(); err != nil {
		if errors.Is(err, ErrNoPlayerFound) {
			return p.conf.NoPlayerStatus, nil
		}
		return "", err
	}

	if err := p.unsafeUpdateMetadata(); err != nil {
		return "", err
	}

	return p.unsafeFormatStatus(), nil
}

func (p *Player) call(method string, args ...any) (any, error) {
	call := p.dbus.Object(
		"org.mpris.MediaPlayer2."+p.conf.Name,
		playerObjectPath,
	).Call(method, 0, args...)

	if call.Err != nil {
		if call.Err.Error() == "The name org.mpris.MediaPlayer2."+
			p.conf.Name+
			" was not provided by any .service files" {
			return nil, ErrNoPlayerFound
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

func (s *Swaypanion) playerHandler(args []string, w io.Writer) error {
	if len(args) == 0 {
		s.player.currentMu.Lock()
		defer s.player.currentMu.Unlock()

		if err := s.player.unsafeUpdatePlaybackStatus(); err != nil {
			if !errors.Is(err, ErrNoPlayerFound) {
				return err
			}

			return s.player.Launch()
		}

		return s.player.Show()
	}

	switch args[0] {
	case "status":
		status, err := s.player.FormattedStatus()
		if err != nil {
			return err
		}

		return writeString(w, status)
	case "playpause":
		return s.player.PlayPause()
	case "previous":
		return s.player.Previous()
	case "next":
		return s.player.Next()
	case "subscribe":
		status, id := s.player.subscriptions.Subscribe()

		for st := range status {
			if err := writeString(w, st); err != nil {
				s.player.subscriptions.Unsubscribe(id)

				if strings.Contains(err.Error(), "broken pipe") {
					return nil
				}

				return fmt.Errorf("failed to send player status to client: %w", err)

			}
		}

		return nil
	default:
		return s.unknownArgument(w, "player", args[0])
	}
}

func (s *Swaypanion) registerPlayer() {
	if s.conf.Player.Disable {
		return
	}

	s.register(
		"player", s.playerHandler, "get or change the audio player status",
		"", "start or show the player",
		"playpause", "play or pause the music",
		"previous", "go back to previous track",
		"next", "skip to next track",
		"subscribe", "receive the player status as soon as it changes",
	)
}
