package swaypanion

import (
	"errors"
	"io"
	"os/exec"
	"strconv"
)

var ErrSwayClientMissing = errors.New("sway client missing")

func writeInt(w io.Writer, i int) error {
	_, err := w.Write([]byte(strconv.Itoa(i)))
	return err
}

func execCmd(sway *SwayClient, commandType CommandType, command string) error {
	switch commandType {
	case CommandTypeExec:
		return exec.Command("sh", "-c", command).Run()
	case CommandTypeSway:
		if sway == nil {
			return ErrSwayClientMissing
		}
		return sway.RunCommand(command)
	}

	return nil
}

// func receiveSwayEvent[T sway.Event](eventType i3.EventType, handleEvent func(T)) func() error {
// 	receiver := i3.Subscribe(eventType)

// 	go func() {
// 		for receiver.Next() {
// 			handleEvent(receiver.Event().(T))
// 		}
// 	}()

// 	return receiver.Close
// }
