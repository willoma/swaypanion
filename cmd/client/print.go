package main

import (
	"os"

	"github.com/willoma/swaypanion/common"
)

func eraseLine() error {
	_, err := os.Stdout.Write([]byte{carriageReturn})
	return err
}

func newLine() {
	if _, err := os.Stdout.Write([]byte{'\n'}); err != nil {
		printError("Failed to print response", err)
	}
}

func printPrompt() {
	if _, err := os.Stdout.WriteString("> "); err != nil {
		printError("Failed to print prompt", err)
	}
}

func printLine(line string) {
	if err := eraseLine(); err != nil {
		printError("Failed to print response", err)
	}

	if _, err := os.Stdout.WriteString(line); err != nil {
		printError("Failed to print response", err)
	}

	newLine()
}

func printError(msg string, err error) {
	eraseLine()

	common.LogError(`/!\ `+msg, err)

	os.Exit(1)
}
