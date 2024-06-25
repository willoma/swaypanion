package swaypanion

import (
	"strconv"
	"strings"

	"github.com/joshuarubin/go-sway"
)

type windowMatchType string

const (
	windowMatchAppID      = "app_id"
	windowMatchPID        = "pid"
	windowMatchShell      = "shell"
	windowMatchWindowID   = "window_id"
	windowMatchTitle      = "title"
	windowMatchClass      = "class"
	windowMatchInstance   = "instance"
	windowMatchWindowRole = "window_role"
	windowMatchWindowType = "window_type"
)

type WindowIdentification struct {
	matchType windowMatchType `yaml:"type"`
	partial   bool            `yaml:"partial"`
	match     string          `yaml:"match"`
}

func (id WindowIdentification) matchWindow(win *sway.Node) bool {
	if win == nil {
		return false
	}

	if id.partial {
		return id.matchPartialWindow(win)
	}

	switch id.matchType {
	case windowMatchAppID:
		return win.AppID != nil && *win.AppID == id.match
	case windowMatchPID:
		pid, err := strconv.ParseUint(id.match, 10, 32)
		if err != nil {
			return false
		}

		return win.PID != nil && *win.PID == uint32(pid)
	case windowMatchShell:
		return win.Shell != nil && *win.Shell == id.match
	case windowMatchWindowID:
		wid, err := strconv.ParseInt(id.match, 10, 64)
		if err != nil {
			return false
		}

		return win.Window != nil && *win.Window == wid
	case windowMatchTitle:
		return win.WindowProperties != nil && win.WindowProperties.Title == id.match
	case windowMatchClass:
		return win.WindowProperties != nil && win.WindowProperties.Class == id.match
	case windowMatchInstance:
		return win.WindowProperties != nil && win.WindowProperties.Instance == id.match
	case windowMatchWindowRole:
		return win.WindowProperties != nil && win.WindowProperties.Role == id.match
	case windowMatchWindowType:
		return win.WindowProperties != nil && win.WindowProperties.Type == id.match
	default:
		return false
	}
}

func (id WindowIdentification) matchPartialWindow(win *sway.Node) bool {
	switch id.matchType {
	case windowMatchAppID:
		return win.AppID != nil && strings.Contains(*win.AppID, id.match)
	case windowMatchPID:
		pid, err := strconv.ParseUint(id.match, 10, 32)
		if err != nil {
			return false
		}

		return win.PID != nil && *win.PID == uint32(pid)
	case windowMatchShell:
		return win.Shell != nil && *win.Shell == id.match
	case windowMatchWindowID:
		wid, err := strconv.ParseInt(id.match, 10, 64)
		if err != nil {
			return false
		}

		return win.Window != nil && *win.Window == wid
	case windowMatchTitle:
		return win.WindowProperties != nil && strings.Contains(win.WindowProperties.Title, id.match)
	case windowMatchClass:
		return win.WindowProperties != nil && strings.Contains(win.WindowProperties.Class, id.match)
	case windowMatchInstance:
		return win.WindowProperties != nil && strings.Contains(win.WindowProperties.Instance, id.match)
	case windowMatchWindowRole:
		return win.WindowProperties != nil && strings.Contains(win.WindowProperties.Role, id.match)
	case windowMatchWindowType:
		return win.WindowProperties != nil && strings.Contains(win.WindowProperties.Type, id.match)
	default:
		return false
	}
}
