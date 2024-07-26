package main

import (
	"bytes"
	"sync"

	socketclient "github.com/willoma/swaypanion/socket/client"
)

type client struct {
	bufMu sync.Mutex
	buf   bytes.Buffer

	client  *socketclient.Client
	stopped chan struct{}
}

func newClient() *client {
	c := &client{
		stopped: make(chan struct{}),
	}

	socketClient, err := socketclient.New()
	if err != nil {
		c.interactivePrintError("Failed to run socket client", err)
	}

	c.client = socketClient

	return c
}

func (c *client) close() {
	if err := c.client.Close(); err != nil {
		c.interactivePrintError("Failed to close socket client", err)
	}
}
