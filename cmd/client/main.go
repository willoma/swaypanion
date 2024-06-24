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
		println("Failed to connect to UNIX socket: ", err.Error())
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
		println("Failed to send command: ", err)
	}

	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		println(scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		println("Failed to read response from UNIX socket: ", err)
	}
}
