package config

type Player struct {
	config[*Player] `yaml:"-"`

	PlayerName string       `yaml:"player_name"`
	Start      Command      `yaml:"start"`
	Show       Command      `yaml:"show"`
	Waybar     WaybarString `yaml:"player"`
}

var DefaultPlayer = &Player{
	PlayerName: "spotify",
	Start: Command{
		Type:    CommandTypeShell,
		Command: "/usr/bin/spotify",
	},
	Show: Command{
		Type:    CommandTypeSway,
		Command: "[instance=spotify] move workspace current, focus",
	},
	Waybar: WaybarString{
		Icons: map[string]string{
			"no player found": " ",
			"Playing":         " ",
			"Paused":          " ",
			"Stopped":         " ",
		},
		FormatText:    "{icon}{artist? }{artist}{title? - }{title}",
		FormatTooltip: "{artists}{album?\n  }{album}{title?\n}{title}",
	},
}

func (p *Player) applyDefault() {
	if p.PlayerName == "" {
		p.PlayerName = DefaultPlayer.PlayerName
	}

	if p.Start.Type == "" {
		p.Start.Type = DefaultPlayer.Start.Type
	}

	if p.Start.Command == "" {
		p.Start.Command = DefaultPlayer.Start.Command
	}

	if p.Show.Type == "" {
		p.Show.Type = DefaultPlayer.Show.Type
	}

	if p.Show.Command == "" {
		p.Show.Command = DefaultPlayer.Show.Command
	}

	p.Waybar = p.Waybar.applyDefault(DefaultPlayer.Waybar)
}
