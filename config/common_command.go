package config

import (
	"errors"
	"os/exec"

	"github.com/willoma/swaypanion/common"
	"github.com/willoma/swaypanion/sway"
)

type CommandType string

const (
	CommandTypeShell CommandType = "shell"
	CommandTypeSway  CommandType = "sway"
)

var (
	ErrNoCommand          = errors.New("no command defined")
	ErrInvalidCommandType = errors.New("invalid command type")
)

type Command struct {
	Type    CommandType `yaml:"cmd_type"`
	Command string      `yaml:"command"`
}

func (c Command) Run(s *sway.Client) error {
	if c.Command == "" {
		return ErrNoCommand
	}

	switch c.Type {
	case CommandTypeShell:
		cmd := exec.Command("sh", "-c", c.Command)

		if err := cmd.Start(); err != nil {
			return err
		}

		if err := cmd.Process.Release(); err != nil {
			return err
		}

		return nil
	case CommandTypeSway:
		return s.RunCommand(c.Command)
	default:
		return common.Errorf(string(c.Type), ErrInvalidCommandType)
	}
}
