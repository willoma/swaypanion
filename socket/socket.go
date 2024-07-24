package socket

import (
	"io"
	"syscall"
)

// Socket message format:
//
//	command - gs - value - gs - complement - gs - complement [...] eor
//
// Single-byte values:
//
//   - gs: [groupSeparator]
//   - eor: [endOfRecord]

const (
	groupSeparator byte = 0x1d
	endOfRecord    byte = 0x1e

	socketName = "swaypanion-dev.sock"
)

const (
	fieldCommand int = iota
	fieldValue
)

var Path string

func init() {
	if socketdir, ok := syscall.Getenv("XDG_RUNTIME_DIR"); ok {
		Path = socketdir + "/" + socketName
	}

	Path = "/tmp/" + socketName
}

type Message struct {
	Command    string
	Value      string
	Complement []string
}

func (m *Message) WriteTo(w io.Writer) (int64, error) {
	if m == nil {
		return 0, nil
	}

	n := 0
	if nn, err := w.Write([]byte(m.Command)); err != nil {
		return int64(n), err
	} else {
		n += nn
	}

	if _, err := w.Write([]byte{groupSeparator}); err != nil {
		return int64(n), err
	} else {
		n += 1
	}

	if nn, err := w.Write([]byte(m.Value)); err != nil {
		return int64(n), err
	} else {
		n += nn
	}

	for _, c := range m.Complement {
		if _, err := w.Write([]byte{groupSeparator}); err != nil {
			return int64(n), err
		} else {
			n += 1
		}

		if nn, err := w.Write([]byte(c)); err != nil {
			return int64(n), err
		} else {
			n += nn
		}
	}

	if _, err := w.Write([]byte{endOfRecord}); err != nil {
		return int64(n), err
	} else {
		n += 1
	}

	return int64(n), nil
}

func (m *Message) ReadFrom(r io.Reader) (n int64, err error) {
	var (
		byteBuf   = make([]byte, 1)
		fieldBuf  = make([]byte, 0, 32)
		num       int64
		fieldStep int
	)

	for {
		if _, err := r.Read(byteBuf); err != nil {
			return num, err
		}
		num++

		switch byteBuf[0] {
		case groupSeparator:
			switch fieldStep {
			case fieldCommand:
				m.Command = string(fieldBuf)
			case fieldValue:
				m.Value = string(fieldBuf)
			default:
				m.Complement = append(m.Complement, string(fieldBuf))
			}

			fieldBuf = make([]byte, 0, 32)
			fieldStep++
		case endOfRecord:
			if len(fieldBuf) > 0 {
				switch fieldStep {
				case fieldCommand:
					m.Command = string(fieldBuf)
				case fieldValue:
					m.Value = string(fieldBuf)
				default:
					m.Complement = append(m.Complement, string(fieldBuf))
				}
			}

			return num, nil
		default:
			fieldBuf = append(fieldBuf, byteBuf[0])
		}
	}
}
