package swaypanion

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net"
	"os"
	"strings"

	"github.com/willoma/swaypanion/socket"
)

func (s *Swaypanion) socketServer() error {
	socketpath := socket.Path()

	if _, err := os.Stat(socketpath); err == nil {
		if rErr := os.Remove(socketpath); rErr != nil {
			return fmt.Errorf("failed to remove previous socket file: %w", rErr)
		}
	} else if !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("failed to check previous socket file: %w", err)
	}

	var err error
	if s.socket, err = net.Listen("unix", socketpath); err != nil {
		return fmt.Errorf("failed to listen to UNIX socket: %w", err)
	}

	go func() {
		for {
			// Accept an incoming connection.
			conn, err := s.socket.Accept()
			if err != nil {
				if errors.Is(err, net.ErrClosed) {
					return
				}

				slog.Error("Failed to accept connection on UNIX socket", "error", err)
			}

			go s.handleSocket(conn)
		}
	}()

	return nil
}

func (s *Swaypanion) handleSocket(conn net.Conn) {
	defer func() {
		if err := conn.Close(); err != nil {
			slog.Error("Failed to close socket", "error", err)
		}
	}()

	input, err := bufio.NewReader(conn).ReadString(socket.EndOfCommand)
	if err != nil {
		slog.Error("Failed to read command", "error", err)
		return
	}

	input = strings.TrimRight(input, string([]byte{socket.EndOfCommand}))

	cmdAndArgs := strings.Split(input, string([]byte{socket.ArgSeparator}))

	command := string(cmdAndArgs[0])

	switch command {
	case "":
		if err := s.commandsList(conn); err != nil {
			slog.Error("Failed to send commands list", "error", err)
		}

		return
	case "help":
		if err := s.help(conn, cmdAndArgs[1:]...); err != nil {
			slog.Error("Failed to send help", "error", err)
		}

		return
	case "config":
		if err := s.dumpConfig(conn); err != nil {
			slog.Error("Failed to send config information", "error", err)

		}

		return
	case "defaultconfig":
		if err := s.dumpDefaultConfig(conn); err != nil {
			slog.Error("Failed to send config information", "error", err)

		}

		return
	}

	reg, ok := s.processors[command]
	if !ok {
		if err := s.unknownCommand(conn, command); err != nil {
			slog.Error("Failed to send response", "error", err)
		}

		return
	}

	if procErr := reg.processor(cmdAndArgs[1:], conn); procErr != nil {
		slog.Error("Failed to process request", "request", strings.ReplaceAll(input, string([]byte{socket.ArgSeparator}), " "), "error", procErr)

		var buf bytes.Buffer

		buf.WriteString(procErr.Error())
		buf.WriteByte('\n')

		if _, wErr := buf.WriteTo(conn); wErr != nil {
			slog.Error("Failed to send response", "error", wErr)
		}
	}
}
