package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/willoma/swaypanion"
)

func main() {
	reloadSignal := make(chan os.Signal, 1)
	signal.Notify(reloadSignal, syscall.SIGHUP)
	stopSignal := make(chan os.Signal, 1)
	signal.Notify(stopSignal, syscall.SIGINT, syscall.SIGTERM)

	slog.Info("Starting swaypanion")

	for {
		s, err := swaypanion.New()
		if err != nil {
			slog.Error("Failed to start swaypanion", "error", err)
			os.Exit(1)
		}

		select {
		case <-stopSignal:
			slog.Info("Stopping swaypanion")
			s.Stop()
			return
		case <-reloadSignal:
			slog.Info("Reloading swaypanion")
			s.Stop()
		}
	}
}
