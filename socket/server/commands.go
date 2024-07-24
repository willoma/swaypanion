package socketserver

import (
	"errors"
	"strings"
)

type Command struct {
	Fn            func(conn *Connection, value string, complement []string)
	Desc          string
	ArgumentsDesc []string
}

type Commands map[string]*Command

const (
	stringFieldName int = iota
	stringFieldDescription
	stringFieldArgumentDescription
)

// NewCommands builds a Commands object with its content.
func NewCommands(args ...any) Commands {
	commands := Commands{}

	var (
		currentName        string
		currentCommand     *Command
		currentStringField int
	)

	for _, arg := range args {
		switch arg := arg.(type) {
		case string:
			switch currentStringField {
			case stringFieldName:
				currentName = arg
				currentStringField++
			case stringFieldDescription:
				currentCommand.Desc = arg
				currentStringField++
			default:
				currentCommand.ArgumentsDesc = append(currentCommand.ArgumentsDesc, arg)
			}
		case func(conn *Connection, value string, complement []string):
			if currentName != "" && currentCommand != nil {
				commands[currentName] = currentCommand
				currentName = ""
			}

			currentStringField = 0
			currentCommand = &Command{
				Fn: arg,
			}
		}
	}

	if currentName != "" && currentCommand != nil {
		commands[currentName] = currentCommand
	}

	return commands
}

func (c Commands) commandByName(name string) (*Command, error) {
	if name == "" {
		return nil, errors.New("no command provided")
	}
	if command, ok := c[name]; ok {
		return command, nil
	}

	fields := strings.Fields(name)

	return c.commandByShortcut(fields)
}

type commandCandidate struct {
	fields  []string
	command *Command
}

func (c Commands) commandByShortcut(req []string) (*Command, error) {
	// Filter out candidates so that we keep only those with the same amount of fields.
	candidates := make(map[string]commandCandidate)
	for cmdName, command := range c {
		fields := strings.Fields(cmdName)
		if len(fields) != len(req) {
			continue
		}

		candidates[cmdName] = commandCandidate{fields, command}
	}

	for i, field := range req {
		for name, candidate := range candidates {
			if !strings.HasPrefix(candidate.fields[i], field) {
				delete(candidates, name)
			}
		}
	}

	switch len(candidates) {
	case 0:
		return nil, errors.New("command not found")
	case 1:
		for _, candidate := range candidates {
			return candidate.command, nil
		}
	default:
		msg := "multiple commands match: "
		for name := range candidates {
			msg += name + ", "
		}
		msg = msg[:len(msg)-2]
		return nil, errors.New(msg)
	}

	return nil, errors.New("unknown error")
}
