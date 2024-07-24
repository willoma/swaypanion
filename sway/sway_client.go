package sway

import (
	"context"
	"time"

	"github.com/joshuarubin/go-sway"
	"github.com/willoma/swaypanion/common"
)

const swayRetryDelay = 2 * time.Second

type Client struct {
	client sway.Client

	ctx    context.Context
	cancel context.CancelFunc
}

func NewClient() *Client {
	ctx, cancel := context.WithCancel(context.Background())

	s := &Client{
		ctx:    ctx,
		cancel: cancel,
	}

	go s.connect()

	return s
}

func (s *Client) connect() {
	var err error

	for {
		s.client, err = sway.New(s.ctx)
		if err == nil {
			break
		}

		if err.Error() == "$SWAYSOCK is empty" {
			common.LogInfo("The SWAYSOCK environment variable is empty, will never connect to Sway")
			break
		}

		common.LogError("Failed to connect to Sway, trying again later", err)

		time.Sleep(swayRetryDelay)
	}
}

func (s *Client) Close() {
	s.cancel()
}

func CountWindows(node *sway.Node) int {
	if len(node.Nodes) == 0 && len(node.FloatingNodes) == 0 {
		return 1
	}

	var count int

	for _, subnode := range node.Nodes {
		count += CountWindows(subnode)
	}

	for _, subnode := range node.FloatingNodes {
		count += CountWindows(subnode)
	}

	return count
}

func hasFocus(node *sway.Node) bool {
	if node.Focused {
		return true
	}

	for _, subnode := range node.Nodes {
		if hasFocus(subnode) {
			return true
		}
	}

	for _, subnode := range node.FloatingNodes {
		if hasFocus(subnode) {
			return true
		}
	}

	return false
}
