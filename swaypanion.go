package swaypanion

import (
	"github.com/willoma/swaypanion/config"
	"github.com/willoma/swaypanion/modules"
	"github.com/willoma/swaypanion/notification"
	socketserver "github.com/willoma/swaypanion/socket/server"
	"github.com/willoma/swaypanion/sway"
)

type Swaypanion struct {
	config *config.Config

	sway         *sway.Client
	notification *notification.Notification
	socketserver *socketserver.Server

	coreNotifier     *notification.MessageNotifier
	stopConfigReload func()

	modules []any
}

func New() (*Swaypanion, error) {
	sock, err := socketserver.New()
	if err != nil {
		return nil, err
	}

	notif, err := notification.New()
	if err != nil {
		return nil, err
	}

	coreNotifier := notif.MessageNotifier()

	conf, err := config.New(coreNotifier)
	if err != nil {
		return nil, err
	}

	s := &Swaypanion{
		config:       conf,
		sway:         sway.NewClient(),
		notification: notif,
		socketserver: sock,
		coreNotifier: coreNotifier,
	}

	s.socketserver.AddCommands(conf.SocketCommands())

	s.register(modules.NewBacklight(conf.Backlight, notif))
	s.register(modules.NewPlayer(conf.Player, s.sway))
	s.register(modules.NewVolume(conf.Volume, notif))
	s.register(modules.NewSwayNodes(conf.SwayNodes, s.sway))

	s.reloadConfig(conf)
	s.stopConfigReload = conf.ListenReload(s.reloadConfig)

	coreNotifier.Notify("Swaypanion started")

	return s, nil
}

func (s *Swaypanion) Stop() error {
	if s.stopConfigReload != nil {
		s.stopConfigReload()
	}

	if err := s.config.Close(); err != nil {
		return err
	}

	for _, module := range s.modules {
		if stopperModule, ok := module.(Stopper); ok {
			stopperModule.Stop()
		}
	}

	s.sway.Close()

	if err := s.socketserver.Close(); err != nil {
		return err
	}

	s.coreNotifier.Notify("Swaypanion stopped")

	return nil
}

func (s *Swaypanion) reloadConfig(conf *config.Config) {
	s.coreNotifier.Reconfigure(conf.CoreMessages)
}

func (s *Swaypanion) register(module any) {
	s.modules = append(s.modules, module)

	if mod, ok := module.(socketCommander); ok {
		s.socketserver.AddCommands(mod.SocketCommands())
	}
}
