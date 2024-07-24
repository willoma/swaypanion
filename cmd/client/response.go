package main

import (
	"strings"

	"github.com/willoma/swaypanion/socket"
)

func printResponse(msg *socket.Message) {
	switch msg.Command {
	case "help":
		if len(msg.Complement) == 0 {
			return
		}

		printHelp(msg.Value, msg.Complement[0], msg.Complement[1:]...)
	default:
		printStandardResponse(msg)
	}
}

func printHelp(command, description string, arguments ...string) {
	switch len(arguments) {
	case 0:
		printLine(command + ": " + description)
	case 1:
		printLine(command + ": " + description + " (Argument: " + arguments[0] + ")")
	default:
		line := command + ": " + description + " (Arguments: " + arguments[0]
		for _, arg := range arguments[1:] {
			line += ", " + arg
		}
		line += ")"

		printLine(line)
	}
}

func printStandardResponse(msg *socket.Message) {
	var separator string
	if strings.Contains(msg.Value, "\n") {
		separator = ":\n"
	} else {
		separator = ": "
	}
	switch len(msg.Complement) {
	case 0:
		printLine(msg.Command + separator + msg.Value)
	case 1:
		printLine(msg.Command + separator + msg.Value + " (" + msg.Complement[0] + ")")
	default:
		line := msg.Command + separator + msg.Value + " (" + msg.Complement[0]
		for _, compl := range msg.Complement[1:] {
			line += ", " + compl
		}
		line += ")"

		printLine(line)
	}
}
