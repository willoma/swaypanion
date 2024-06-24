package swaypanion

import (
	"bytes"
	"io"
)

func (s *Swaypanion) unknownCommand(w io.Writer, command string) error {
	var buf bytes.Buffer

	buf.WriteString("Command unknown: ")
	buf.WriteString(command)
	buf.WriteString("\n\n")

	if _, err := buf.WriteTo(w); err != nil {
		return err
	}

	return s.commandsList(w)
}

func (s *Swaypanion) commandsList(w io.Writer) error {
	var buf bytes.Buffer

	buf.WriteString("Known commands:\n- help: show commands and arguments\n")

	for name, reg := range s.processors {
		buf.WriteString("- ")
		buf.WriteString(name)

		if reg.description != "" {
			buf.WriteString(": ")
			buf.WriteString(reg.description)
		}

		buf.WriteByte('\n')
	}

	_, err := buf.WriteTo(w)
	return err
}

func (s *Swaypanion) unknownArgument(w io.Writer, command, argument string) error {
	var buf bytes.Buffer

	buf.WriteString("Argument unknown for command ")
	buf.WriteString(command)
	buf.WriteString(": ")
	buf.WriteString(argument)
	buf.WriteString("\n\n")

	if _, err := buf.WriteTo(w); err != nil {
		return err
	}

	return s.help(w, command)
}

func (s *Swaypanion) help(w io.Writer, args ...string) error {
	if len(args) == 0 {
		return s.commandsList(w)
	}

	reg, ok := s.processors[args[0]]
	if !ok {
		return s.unknownCommand(w, args[0])
	}

	return processorHelp(w, args[0], reg)
}

func processorHelp(w io.Writer, command string, reg registeredProcessor) error {
	var buf bytes.Buffer

	for _, arg := range reg.arguments {
		buf.WriteString(command)

		if arg.name != "" {
			buf.WriteByte(' ')
			buf.WriteString(arg.name)
		}

		buf.WriteString(": ")
		buf.WriteString(arg.description)
		buf.WriteByte('\n')

	}

	_, err := buf.WriteTo(w)
	return err
}
