package main

import (
	"bufio"
	"bytes"
	"net"
	"os"

	"github.com/willoma/swaypanion/socket"
)

func main() {
	socketpath := socket.Path()

	conn, err := net.Dial("unix", socketpath)
	if err != nil {
		printError("Failed to connect to UNIX socket", err)
	}

	defer conn.Close()

	var buf bytes.Buffer

	if len(os.Args[1:]) > 0 {
		buf.WriteString(os.Args[1])

		for _, arg := range os.Args[2:] {
			buf.WriteByte(socket.ArgSeparator)
			buf.WriteString(arg)
		}
	}

	buf.WriteByte(socket.EndOfCommand)

	if _, err := buf.WriteTo(conn); err != nil {
		printError("Failed to send command", err)
	}

	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		if _, err := os.Stdout.Write(append(scanner.Bytes(), '\n')); err != nil {
			printError("Failed to print to output", err)
		}
	}

	if err := scanner.Err(); err != nil {
		printError("Failed to read response from UNIX socket", err)
	}
}

func printError(msg string, err error) {
	os.Stderr.WriteString(msg + ": " + err.Error() + "\n")
	os.Exit(1)
}
