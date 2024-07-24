package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/willoma/swaypanion"
	"github.com/willoma/swaypanion/common"
)

func main() {
	reloadSignal := make(chan os.Signal, 1)
	signal.Notify(reloadSignal, syscall.SIGHUP)
	stopSignal := make(chan os.Signal, 1)
	signal.Notify(stopSignal, syscall.SIGINT, syscall.SIGTERM)

	common.LogInfo("Starting swaypanion")

	for {
		s, err := swaypanion.New()
		if err != nil {
			common.LogError("Failed to start swaypanion", err)
			os.Exit(1)
		}

		select {
		case <-stopSignal:
			common.LogInfo("Stopping swaypanion")
			s.Stop()
			return
		case <-reloadSignal:
			common.LogInfo("Reloading swaypanion")
			s.Stop()
		}
	}
}
