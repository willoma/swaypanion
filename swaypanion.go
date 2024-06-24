package swaypanion

import (
	"io"
	"log/slog"
	"net"
)

type processor func(args []string, w io.Writer) error

type registeredProcessor struct {
	processor   processor
	description string
	arguments   []struct {
		name        string
		description string
	}
}

type Swaypanion struct {
	conf *config

	sway *SwayClient

	notification *Notification
	backlight    *Backlight
	player       *Player
	volume       *Volume
	workspaces   *Workspaces

	socket net.Listener

	processors map[string]registeredProcessor
}

func (s *Swaypanion) registerProcessors() {
	s.processors = map[string]registeredProcessor{}

	s.registerBacklight()
	s.registerPlayer()
	s.registerVolume()
	s.registerWorkspaces()
}

func New() (*Swaypanion, error) {
	var (
		s   = &Swaypanion{}
		err error
	)

	if s.conf, err = newConfig(); err != nil {
		return nil, err
	}

	if s.sway, err = NewSwayClient(); err != nil {
		return nil, err
	}

	if s.notification, err = NewNotification(s.conf.Notification); err != nil {
		return nil, err
	}

	if s.backlight, err = NewBacklight(s.conf.Backlight); err != nil {
		return nil, err
	}

	if s.player, err = NewPlayer(s.conf.Player, s.sway); err != nil {
		return nil, err
	}

	if s.volume, err = NewVolume(s.conf.Volume); err != nil {
		return nil, err
	}

	if s.workspaces, err = NewWorkspaces(s.conf.Workspaces, s.sway); err != nil {
		return nil, err
	}

	s.registerProcessors()

	err = s.socketServer()

	return s, err
}

func (s *Swaypanion) Stop() {
	if s.volume != nil {
		if err := s.volume.Close(); err != nil {
			slog.Error("Failed to close volume connection", "error", err)
		}
	}

	if s.socket != nil {
		if err := s.socket.Close(); err != nil {
			slog.Error("Failed to close socket listener", "error", err)
		}
	}

	if s.sway != nil {
		s.sway.Close()
	}
}

func (s *Swaypanion) register(name string, proc processor, description string, argsAndDescriptions ...string) {
	r := registeredProcessor{
		processor:   proc,
		description: description,
	}
	for i := 0; i < len(argsAndDescriptions); i += 2 {
		r.arguments = append(r.arguments, struct {
			name        string
			description string
		}{
			argsAndDescriptions[i], argsAndDescriptions[i+1],
		})
	}

	s.processors[name] = r
}
