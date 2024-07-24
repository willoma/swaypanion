//go:build slog

package common

import "log/slog"

func LogError(msg string, err error) {
	slog.Error(msg, "error", err)
}

func LogInfo(msg string) {
	slog.Info(msg)
}
