package config

import (
	"gopkg.in/yaml.v3"

	"github.com/willoma/swaypanion/common"
	socketserver "github.com/willoma/swaypanion/socket/server"
)

func (c *Config) SocketCommands() socketserver.Commands {
	return socketserver.NewCommands(
		c.socketConfig, "configuration", "Dump current configuration",
		c.socketConfigDefault, "configuration default", "Dump default configuration",
	)
}

func (c *Config) socketConfig(conn *socketserver.Connection, _ string, _ []string) {
	dump, err := yaml.Marshal(c)
	if err != nil {
		common.LogError("Failed to dump current configuration", err)
		conn.SendError("failed to dump current configuration")

		return
	}

	conn.SendString("configuration", string(dump))
}

func (c *Config) socketConfigDefault(conn *socketserver.Connection, _ string, _ []string) {
	dump, err := yaml.Marshal(Default)
	if err != nil {
		common.LogError("Failed to dump default configuration", err)
		conn.SendError("failed to dump default configuration")

		return
	}

	conn.SendString("configuration default", string(dump))
}
