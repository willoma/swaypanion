package main

import (
	"os"
	"strings"

	"github.com/willoma/swaypanion/socket"
)

func (c *client) directPrintError(msg string, err error) {
	c.printMu.Lock()
	defer c.printMu.Unlock()

	c.unsafePrintError(msg, err)

}
func (c *client) interactivePrintError(msg string, err error) {
	c.printMu.Lock()
	defer c.printMu.Unlock()

	c.buf.WriteString("\r/!\\ ")

	c.unsafePrintError(msg, err)

}

func (c *client) unsafePrintError(msg string, err error) {
	c.buf.WriteString(msg)
	c.buf.WriteString(": ")
	c.buf.WriteString(err.Error())
	c.buf.WriteByte('\n')

	c.buf.WriteTo(os.Stderr)

	os.Exit(1)
}

func (c *client) printLine(l string) {
	c.printMu.Lock()
	defer c.printMu.Unlock()

	c.buf.WriteString(l)
	c.buf.WriteByte('\n')

	c.buf.WriteTo(os.Stdout)
}

func (c *client) interactivePrintPrompt() {
	c.printMu.Lock()
	defer c.printMu.Unlock()

	os.Stdout.WriteString("> ")
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

func (c *client) directPrintResponseCommon(msg *socket.Message) {
	c.printMu.Lock()
	defer c.printMu.Unlock()

	c.unsafePrintResponseCommon(msg)

	c.buf.WriteTo(os.Stdout)
}

func (c *client) interactivePrintResponseCommon(msg *socket.Message) {
	c.printMu.Lock()
	defer c.printMu.Unlock()

	c.buf.WriteByte('\r')
	c.buf.WriteString(msg.Command)

	if strings.Contains(msg.Value, "\n") {
		c.buf.WriteString(":\n")
	} else {
		c.buf.WriteString(": ")
	}

	c.unsafePrintResponseCommon(msg)

	c.buf.WriteString("> ")

	c.buf.WriteTo(os.Stdout)
}

func (c *client) unsafePrintResponseCommon(msg *socket.Message) {
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
}

func (c *client) directPrintResponseHelp(command, description string, arguments []string) {
	c.printMu.Lock()
	defer c.printMu.Unlock()

	c.unsafePrintResponseHelp(command, description, arguments)

	c.buf.WriteTo(os.Stdout)
}

func (c *client) interactivePrintResponseHelp(command, description string, arguments []string) {
	c.printMu.Lock()
	defer c.printMu.Unlock()

	c.buf.WriteByte('\r')

	c.unsafePrintResponseHelp(command, description, arguments)

	c.buf.WriteString("> ")

	c.buf.WriteTo(os.Stdout)
}

func (c *client) unsafePrintResponseHelp(command, description string, arguments []string) {
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
}
