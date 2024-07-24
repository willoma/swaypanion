//go:build !slog

package common

import "os"

func LogError(msg string, err error) {
	if err != nil {
		os.Stderr.WriteString(msg + ": " + err.Error() + "\n")
	} else {
		os.Stderr.WriteString(msg + "\n")
	}
}

func LogInfo(msg string) {
	os.Stdout.WriteString(msg + "\n")
}
