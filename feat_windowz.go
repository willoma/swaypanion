package swaypanion

import (
	"io"
)

const (
	defaultWindowsHide1MatchType = windowMatchInstance
	defaultWindowsHide1Partial   = false
	defaultWindowsHide1Match     = "spotify"
)

type WindowsConfig struct {
	Disable      bool                   `yaml:"disable"`
	HideNotClose []WindowIdentification `yaml:"hide_not_close"`
}

func (c WindowsConfig) withDefaults() WindowsConfig {
	if c.HideNotClose == nil {
		c.HideNotClose = []WindowIdentification{
			{defaultWindowsHide1MatchType, defaultWindowsHide1Partial, defaultWindowsHide1Match},
		}
	}

	return c
}

type Windows struct {
	conf WindowsConfig
	sway *SwayClient
}

func NewWindows(conf WindowsConfig, sway *SwayClient) (*Windows, error) {
	if conf.Disable {
		return nil, nil
	}

	return &Windows{conf, sway}, nil
}

func (w *Windows) HideOrCloseCurrentWindow() error {
	window, err := w.sway.focusedNode()
	if err != nil {
		return err
	}

	for _, id := range w.conf.HideNotClose {
		if id.matchWindow(window) {
			return w.sway.RunCommand("move to scratchpad")
		}
	}

	return w.sway.RunCommand("kill")
}

func (s *Swaypanion) windowHandler(args []string, w io.Writer) error {
	if len(args) == 0 {
		return missingArgument("action")
	}

	switch args[0] {
	case "hide_or_close":
		return s.windows.HideOrCloseCurrentWindow()
	default:
		return s.unknownArgument(w, "window", args[0])

	}
}

func (s *Swaypanion) registerWindows() {
	if s.conf.Windows.Disable {
		return
	}

	s.register(
		"window", s.windowHandler, "manipulate the focused window",
		"hide_or_close", "hide or kill the focused window, according to the hide_not_close configuration",
	)
}
