package config

import (
	"strconv"
	"strings"

	"github.com/joshuarubin/go-sway"
)

type WindowMatchType string

const (
	WindowMatchAppID      WindowMatchType = "app_id"
	WindowMatchPID        WindowMatchType = "pid"
	WindowMatchShell      WindowMatchType = "shell"
	WindowMatchWindowID   WindowMatchType = "window_id"
	WindowMatchTitle      WindowMatchType = "title"
	WindowMatchClass      WindowMatchType = "class"
	WindowMatchInstance   WindowMatchType = "instance"
	WindowMatchWindowRole WindowMatchType = "window_role"
	WindowMatchWindowType WindowMatchType = "window_type"
)

type WindowIdentification struct {
	Type    WindowMatchType `yaml:"type"`
	Partial bool            `yaml:"partial"`
	Match   string          `yaml:"match"`
}

func (w WindowIdentification) MatchWindow(win *sway.Node) bool {
	if win == nil {
		return false
	}

	if w.Partial {
		return w.matchPartialWindow(win)
	}

	return w.matchCompleteWindow(win)

}

func (w WindowIdentification) matchCompleteWindow(win *sway.Node) bool {
	switch w.Type {
	case WindowMatchAppID:
		return win.AppID != nil && *win.AppID == w.Match
	case WindowMatchPID:
		pid, err := strconv.ParseUint(w.Match, 10, 32)
		if err != nil {
			return false
		}

		return win.PID != nil && *win.PID == uint32(pid)
	case WindowMatchShell:
		return win.Shell != nil && *win.Shell == w.Match
	case WindowMatchWindowID:
		wid, err := strconv.ParseInt(w.Match, 10, 64)
		if err != nil {
			return false
		}

		return win.Window != nil && *win.Window == wid
	case WindowMatchTitle:
		return win.WindowProperties != nil && win.WindowProperties.Title == w.Match
	case WindowMatchClass:
		return win.WindowProperties != nil && win.WindowProperties.Class == w.Match
	case WindowMatchInstance:
		return win.WindowProperties != nil && win.WindowProperties.Instance == w.Match
	case WindowMatchWindowRole:
		return win.WindowProperties != nil && win.WindowProperties.Role == w.Match
	case WindowMatchWindowType:
		return win.WindowProperties != nil && win.WindowProperties.Type == w.Match
	default:
		return false
	}
}

func (w WindowIdentification) matchPartialWindow(win *sway.Node) bool {
	switch w.Type {
	case WindowMatchAppID:
		return win.AppID != nil && strings.Contains(*win.AppID, w.Match)
	case WindowMatchPID:
		pid, err := strconv.ParseUint(w.Match, 10, 32)
		if err != nil {
			return false
		}

		return win.PID != nil && *win.PID == uint32(pid)
	case WindowMatchShell:
		return win.Shell != nil && *win.Shell == w.Match
	case WindowMatchWindowID:
		wid, err := strconv.ParseInt(w.Match, 10, 64)
		if err != nil {
			return false
		}

		return win.Window != nil && *win.Window == wid
	case WindowMatchTitle:
		return win.WindowProperties != nil && strings.Contains(win.WindowProperties.Title, w.Match)
	case WindowMatchClass:
		return win.WindowProperties != nil && strings.Contains(win.WindowProperties.Class, w.Match)
	case WindowMatchInstance:
		return win.WindowProperties != nil && strings.Contains(win.WindowProperties.Instance, w.Match)
	case WindowMatchWindowRole:
		return win.WindowProperties != nil && strings.Contains(win.WindowProperties.Role, w.Match)
	case WindowMatchWindowType:
		return win.WindowProperties != nil && strings.Contains(win.WindowProperties.Type, w.Match)
	default:
		return false
	}
}
