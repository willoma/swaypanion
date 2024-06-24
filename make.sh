#!/bin/sh -e

GO() {
	CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" "$@"
}

mkdir -p dist/bin

GO -o dist/bin/swaypanion cmd/daemon/main.go
GO -o dist/bin/swaypanionc cmd/client/main.go

if [ "$1" = "install" ]
then
	sudo cp dist/bin/* /usr/local/bin/
fi
