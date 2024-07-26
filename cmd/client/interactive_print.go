package main

import (
	"os"
	"strings"

	"github.com/willoma/swaypanion/socket"
)

const (
	prompt        = "> "
	newlinePrompt = "\n" + prompt
)

func (c *client) interactivePrintError(msg string, err error) {
	c.bufMu.Lock()
	defer c.bufMu.Unlock()

	c.buf.WriteString("\r/!\\ ")
	c.buf.WriteString(msg)
	c.buf.WriteString(": ")
	c.buf.WriteString(err.Error())
	c.buf.WriteString(newlinePrompt)

	c.buf.WriteTo(os.Stderr)

	os.Exit(1)
}

func (c *client) interactivePrintLine(l string) {
	c.bufMu.Lock()
	defer c.bufMu.Unlock()

	c.buf.WriteByte('\n')
	c.buf.WriteString(l)
	c.buf.WriteString(newlinePrompt)

	c.buf.WriteTo(os.Stdout)
}

func (c *client) interactivePrintNewline() {
	c.bufMu.Lock()
	defer c.bufMu.Unlock()

	c.buf.WriteByte('\n')

	c.buf.WriteTo(os.Stdout)
}

func (c *client) interactivePrintPrompt() {
	c.bufMu.Lock()
	defer c.bufMu.Unlock()

	c.buf.WriteString(prompt)

	c.buf.WriteTo(os.Stdout)
}

func (c *client) interactivePrintResponse(msg *socket.Message) {
	switch msg.Command {
	case "help":
		if len(msg.Complement) == 0 {
			return
		}

		c.interactivePrintResponseHelp(msg.Value, msg.Complement[0], msg.Complement[1:])
	default:
		c.interactivePrintResponseCommon(msg)
	}

}

func (c *client) interactivePrintResponseCommon(msg *socket.Message) {
	c.bufMu.Lock()
	defer c.bufMu.Unlock()

	c.buf.WriteByte('\r')
	c.buf.WriteString(msg.Command)

	if strings.Contains(msg.Value, "\n") {
		c.buf.WriteString(":\n")
	} else {
		c.buf.WriteString(": ")
	}

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

	c.buf.WriteString(newlinePrompt)

	c.buf.WriteTo(os.Stdout)
}

func (c *client) interactivePrintResponseHelp(command, description string, arguments []string) {
	c.bufMu.Lock()
	defer c.bufMu.Unlock()

	c.buf.WriteByte('\r')
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

	c.buf.WriteString(newlinePrompt)

	c.buf.WriteTo(os.Stdout)
}

func (c *client) printString(msg string) {
	c.bufMu.Lock()
	defer c.bufMu.Unlock()

	c.buf.WriteByte('\r')
	c.buf.WriteString(msg)
	c.buf.WriteString(newlinePrompt)

	c.buf.WriteTo(os.Stdout)
}
