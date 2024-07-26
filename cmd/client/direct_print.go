package main

import (
	"os"

	"github.com/willoma/swaypanion/socket"
)

func (c *client) directPrintError(msg string, err error) {
	c.bufMu.Lock()
	defer c.bufMu.Unlock()

	c.buf.WriteString(msg)
	c.buf.WriteString(": ")
	c.buf.WriteString(err.Error())
	c.buf.WriteByte('\n')

	c.buf.WriteTo(os.Stderr)

	os.Exit(1)
}

func (c *client) directPrintResponse(msg *socket.Message) {
	switch msg.Command {
	case "help":
		if len(msg.Complement) == 0 {
			return
		}

		c.directPrintResponseHelp(msg.Value, msg.Complement[0], msg.Complement[1:])
	default:
		c.directPrintResponseCommon(msg)
	}

}

func (c *client) directPrintResponseCommon(msg *socket.Message) {
	c.bufMu.Lock()
	defer c.bufMu.Unlock()

	c.buf.WriteString(msg.Value)

	if len(msg.Complement) > 0 {
		c.buf.WriteString(" (")
		c.buf.WriteString(msg.Complement[0])

		if len(msg.Complement) > 1 {
			for _, compl := range msg.Complement[1:] {
				c.buf.WriteString(", ")
				c.buf.WriteString(compl)
			}
		}

		c.buf.WriteByte(')')
	}

	c.buf.WriteByte('\n')

	c.buf.WriteTo(os.Stdout)
}

func (c *client) directPrintResponseHelp(command, description string, arguments []string) {
	c.bufMu.Lock()
	defer c.bufMu.Unlock()

	c.buf.WriteString(command)
	c.buf.WriteString(": ")
	c.buf.WriteString(description)

	if len(arguments) == 1 {
		c.buf.WriteString(" (Argument: ")
		c.buf.WriteString(arguments[0])
		c.buf.WriteByte(')')
	} else if len(arguments) > 1 {
		c.buf.WriteString(" (Arguments: ")
		c.buf.WriteString(arguments[0])
		for _, arg := range arguments[1:] {
			c.buf.WriteString(", ")
			c.buf.WriteString(arg)
		}
		c.buf.WriteByte(')')
	}

	c.buf.WriteByte('\n')

	c.buf.WriteTo(os.Stdout)
}
