package sway

import "errors"

func (c *Client) RunCommand(command string) error {
	resp, err := c.client.RunCommand(c.ctx, command)
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
