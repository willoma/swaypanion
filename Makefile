.PHONY: swaypanion swaypanionc swaypanion-waybar

bin: swaypanion swaypanionc swaypanion-waybar

dist:
	mkdir -p dist

swaypanion: dist
	CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -tags slog -o dist/swaypanion cmd/daemon/*.go

swaypanionc:
	CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o dist/swaypanionc cmd/client/*.go

swaypanion-waybar:
	CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o dist/swaypanion-waybar cmd/waybar/*.go

deb: swaypanion swaypanionc swaypanion-waybar
	mkdir -p dist/pkg/swaypanion/DEBIAN
	cp assets/debian.control dist/pkg/swaypanion/DEBIAN/control
	cp assets/debian.postinst dist/pkg/swaypanion/DEBIAN/postinst

	mkdir -p dist/pkg/swaypanion/usr/bin
	cp dist/swaypanion dist/swaypanionc dist/swaypanion-waybar dist/pkg/swaypanion/usr/bin/

	mkdir -p dist/pkg/swaypanion/usr/lib/systemd/user
	cp assets/swaypanion.service dist/pkg/swaypanion/usr/lib/systemd/user/

	dpkg-deb --build --root-owner-group dist/pkg/swaypanion

clean:
	rm -fr dist
