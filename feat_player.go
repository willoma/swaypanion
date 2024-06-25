package swaypanion

import (
	"errors"
	"html"
	"io"
	"strings"

	"github.com/godbus/dbus/v5"
)

const (
	playerMessagePrefix = "org.mpris.MediaPlayer2.Player"

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
}

func NewPlayer(conf PlayerConfig, sway *SwayClient) (*Player, error) {

	dbusConn, err := dbus.SessionBus()
	if err != nil {
		return nil, err
	}

	return &Player{
		conf: conf,
		sway: sway,
		dbus: dbusConn,
	}, nil
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

func (p *Player) Metadata() (playerMetadata, error) {
	result := playerMetadata{}

	response, err := p.call("org.freedesktop.DBus.Properties.Get", playerMessagePrefix, "Metadata")
	if err != nil {
		return result, err
	}

	if array, ok := response.(map[string]dbus.Variant); ok {
		for k, v := range array {
			switch k {
			case "xesam:album":
				result.album = v.Value().(string)
			case "xesam:artist":
				result.artists = v.Value().([]string)
			case "xesam:title":
				result.title = v.Value().(string)
			}
		}
	}

	return result, nil
}

func (p *Player) PlaybackStatus() (string, error) {
	response, err := p.call("org.freedesktop.DBus.Properties.Get", playerMessagePrefix, "PlaybackStatus")
	if err != nil {
		return "", err
	}

	if v, ok := response.(string); ok {
		return v, nil
	}

	return "", nil
}

func (p *Player) FormattedStatus() (string, error) {
	status, err := p.PlaybackStatus()
	if err != nil {
		if errors.Is(err, ErrNoPlayerFound) {
			return p.conf.NoPlayerStatus, nil
		}
		return "", err
	}

	data, err := p.Metadata()
	if err != nil {
		return "", err
	}

	var artist string
	if len(data.artists) > 0 {
		artist = data.artists[0]
	}
	artists := strings.Join(data.artists, ", ")

	if *p.conf.EscapeHTML {
		return strings.NewReplacer(
			"<status>", status,
			"<title>", html.EscapeString(data.title),
			"<artist>", html.EscapeString(artist),
			"<artists>", html.EscapeString(artists),
			"<album>", html.EscapeString(data.album),
			"\\n", "\n",
			"\\r", "\r",
		).Replace(p.conf.StatusFormat), nil
	} else {
		return strings.NewReplacer(
			"<status>", status,
			"<title>", data.title,
			"<artist>", artist,
			"<artists>", artists,
			"<album>", data.album,
			"\\n", "\n",
			"\\r", "\r",
		).Replace(p.conf.StatusFormat), nil
	}
}

func (p *Player) call(method string, args ...any) (any, error) {
	call := p.dbus.Object(
		"org.mpris.MediaPlayer2."+p.conf.Name,
		"/org/mpris/MediaPlayer2",
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
		if _, err := s.player.PlaybackStatus(); err != nil {
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
	)
}
