package main

import (
	"bytes"
	"io"

	"github.com/willoma/swaypanion/socket"
)

func readMessage(r io.ByteReader) (*socket.Message, error) {
	var (
		current      []byte
		in           byte
		currentField int
		message      = &socket.Message{}
		err          error
	)

	assignCurrent := func() {
		if current == nil {
			return
		}

		content := string(bytes.Join(bytes.Fields(current), []byte{' '}))

		switch currentField {
		case 0:
			message.Command = content
		case 1:
			message.Value = content
		default:
			message.Complement = append(message.Complement, content)
		}

		current = nil
		currentField++
	}

	for {
		if in, err = r.ReadByte(); err != nil {
			if err == io.EOF {
				assignCurrent()
			}

			return message, err
		}

		switch in {
		case ':':
			assignCurrent()
		case '\n':
			assignCurrent()
			return message, nil
		default:
			current = append(current, in)
		}
	}
}
