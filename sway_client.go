package swaypanion

import (
	"context"
	"errors"

	"github.com/joshuarubin/go-sway"
)

type SwayClient struct {
	client sway.Client

	ctx   context.Context
	Close context.CancelFunc
}

func NewSwayClient() (*SwayClient, error) {
	ctx, cancel := context.WithCancel(context.Background())

	client, err := sway.New(ctx)
	if err != nil {
		cancel()
		return nil, err
	}

	return &SwayClient{
		client: client,
		ctx:    ctx,
		Close:  cancel,
	}, nil
}

func (s *SwayClient) GetTree() (*sway.Node, error) {
	return s.client.GetTree(s.ctx)
}

func (s *SwayClient) RunCommand(command string) error {
	resp, err := s.client.RunCommand(s.ctx, command)
	if err != nil {
		return err
	}

	for _, r := range resp {
		if !r.Success {
			return errors.New(r.Error)
		}
	}

	return nil
}
