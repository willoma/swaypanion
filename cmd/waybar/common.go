package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"

	"github.com/willoma/swaypanion/common"
	"github.com/willoma/swaypanion/config"
	"github.com/willoma/swaypanion/socket"
	socketclient "github.com/willoma/swaypanion/socket/client"
)

var jsonReplacer = strings.NewReplacer(
	"\n", "\r",
	"\\r", "\r",
	"\\n", "\r",
	`"`, `\"`,
)

func printError(msg string, err error) {
	common.LogError(msg, err)

	os.Exit(1)
}

func receive(client *socketclient.Client, command string, printEvent func(*socket.Message)) {
	if err := client.Send(&socket.Message{
		Command: command,
	}); err != nil {
		printError("Failed to subscribe", err)
	}

	for {
		msg, err := client.Read()
		if err != nil {
			if !errors.Is(err, io.EOF) {
				printError("Failed to read response", err)
			}

			return
		}

		printEvent(msg)
	}
}

func intPrinter(conf config.WaybarPercent) func(msg *socket.Message) {
	return func(msg *socket.Message) {
		printJSON(conf.FormatValue(msg.Value))
	}
}

func printJSON(alt, text, tooltip string, disabled bool) {
	b := &bytes.Buffer{}

	b.WriteByte('{')

	var hasPrevious bool

	if alt != "" {
		b.WriteString(`"alt":"`)
		jsonReplacer.WriteString(b, alt)
		b.WriteByte('"')
		hasPrevious = true
	}

	if text != "" {
		if hasPrevious {
			b.WriteByte(',')
		}
		b.WriteString(`"text":"`)
		jsonReplacer.WriteString(b, text)
		b.WriteByte('"')
		hasPrevious = true
	}

	if tooltip != "" {
		if hasPrevious {
			b.WriteByte(',')
		}
		b.WriteString(`"tooltip":"`)
		jsonReplacer.WriteString(b, tooltip)
		b.WriteByte('"')
		hasPrevious = true
	}

	if disabled {
		if hasPrevious {
			b.WriteByte(',')
		}
		b.WriteString(`"class":"disabled"`)
	}

	b.WriteString("}\n")

	b.WriteTo(os.Stdout)
}
