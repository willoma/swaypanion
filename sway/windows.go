package sway

import (
	"errors"

	"github.com/joshuarubin/go-sway"
)

var ErrCurrentWindowNotFound = errors.New("current window not found")

func (c *Client) FocusedNode() (*sway.Node, error) {
	root, err := c.client.GetTree(c.ctx)
	if err != nil {
		return nil, err
	}

	focused := root.FocusedNode()

	if focused == nil {
		return nil, ErrCurrentWindowNotFound
	}

	return focused, nil

}
