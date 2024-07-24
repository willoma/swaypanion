package modules

import socketserver "github.com/willoma/swaypanion/socket/server"

const (
	dynworkspaceLabel = "dynworkspace"
	windowLabel       = "window"
)

func (s *SwayNodes) SocketCommands() socketserver.Commands {
	return socketserver.NewCommands(
		s.socketNext, dynworkspaceLabel+" next", "Go to next workspace, creating it if needed",
		s.socketPrev, dynworkspaceLabel+" previous", "Go to previous workspace, creating it if needed",
		s.socketMoveNext, dynworkspaceLabel+" move next", "Move focused window to next workspace, creating it if needed",
		s.socketMovePrev, dynworkspaceLabel+" move previous", "Move focused window to previous workspace, creating it if needed",
		s.socketHideOrClose, windowLabel+" hide-or-close", "Hide or close the focused window, depending on the Swaypanion configuration",
	)
}

func (s *SwayNodes) socketNext(*socketserver.Connection, string, []string) {
	s.next()
}

func (s *SwayNodes) socketPrev(*socketserver.Connection, string, []string) {
	s.prev()
}

func (s *SwayNodes) socketMoveNext(*socketserver.Connection, string, []string) {
	s.moveNext()
}

func (s *SwayNodes) socketMovePrev(*socketserver.Connection, string, []string) {
	s.movePrev()
}

func (s *SwayNodes) socketHideOrClose(*socketserver.Connection, string, []string) {
	s.hideOrClose()
}
