package socketserver

import (
	"errors"
	"io"
	"io/fs"
	"net"
	"os"
	"slices"
	"strings"

	"github.com/willoma/swaypanion/common"
	"github.com/willoma/swaypanion/socket"
)

type Server struct {
	commands Commands
	socket   net.Listener
}

func New() (*Server, error) {
	if _, err := os.Stat(socket.Path); err == nil {
		if rErr := os.Remove(socket.Path); rErr != nil {
			return nil, common.Errorf("failed to remove previous socket file", rErr)
		}
	} else if !errors.Is(err, fs.ErrNotExist) {
		return nil, common.Errorf("failed to check previous socket file", err)
	}

	socket, err := net.Listen("unix", socket.Path)
	if err != nil {
		return nil, common.Errorf("failed to listen to UNIX socket", err)
	}

	s := &Server{commands: Commands{}, socket: socket}

	go s.listen()

	return s, nil
}

func (s *Server) AddCommands(commands Commands) {
	for name, action := range commands {
		s.commands[name] = action
	}
}

func (s *Server) Close() error {
	return s.socket.Close()
}

func (s *Server) listen() {
	for {
		// Accept an incoming connection.
		conn, err := s.socket.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}

			common.LogError("Failed to accept connection on UNIX socket", err)
		}

		go s.handle(conn)
	}
}

func (s *Server) handle(netConn io.ReadWriteCloser) {
	conn := &Connection{conn: netConn}

	defer func() {
		if err := conn.Close(); err != nil {
			if errors.Is(err, net.ErrClosed) {
				// Already closed, no need to close again
				return
			}

			common.LogError("Failed to close socket", err)
		}
	}()

	for {
		msg := conn.Receive()
		if msg == nil {
			return
		}

		s.handleMessage(conn, msg)
	}
}

func (s *Server) handleMessage(conn *Connection, msg *socket.Message) {
	if msg.Command == "close" {
		if err := conn.Close(); err != nil {
			common.LogError("Failed to close connection", err)
		}

		return
	}

	if msg.Command == "help" || msg.Command == "h" || msg.Command == "?" {
		s.writeHelp(conn)
		return
	}

	cmd, err := s.commands.commandByName(msg.Command)
	if err != nil {
		conn.SendError(err.Error(), msg.Command)
		return
	}

	cmd.Fn(conn, msg.Value, msg.Complement)
}

func (s *Server) writeHelp(conn *Connection) {
	msgs := make([]socket.Message, len(s.commands))
	i := 0
	for name, command := range s.commands {
		msgs[i] = socket.Message{
			Command:    "help",
			Value:      name,
			Complement: append([]string{command.Desc}, command.ArgumentsDesc...),
		}
		i++
	}

	slices.SortFunc(msgs, func(a, b socket.Message) int {
		return strings.Compare(a.Value, b.Value)
	})

	msgs = append([]socket.Message{
		{
			Command:    "help",
			Value:      "help",
			Complement: []string{"Show this help message"},
		},
		{
			Command:    "help",
			Value:      "close",
			Complement: []string{"Require the server to close the socket"},
		},
	}, msgs...)

	for _, msg := range msgs {
		conn.Send(msg)
	}
}
