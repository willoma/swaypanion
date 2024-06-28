.PHONY: swaypanion swaypanionc

swaypanion:
	CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o swaypanion cmd/daemon/main.go

swaypanionc:
	CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o swaypanionc cmd/client/main.go

deb: swaypanion swaypanionc
	mkdir -p pkg/swaypanion/DEBIAN
	cp assets/debian.control pkg/swaypanion/DEBIAN/control
	cp assets/debian.postinst pkg/swaypanion/DEBIAN/postinst

	mkdir -p pkg/swaypanion/usr/bin
	cp swaypanion swaypanionc pkg/swaypanion/usr/bin/

	mkdir -p pkg/swaypanion/usr/lib/systemd/user
	cp assets/swaypanion.service pkg/swaypanion/usr/lib/systemd/user/

	dpkg-deb --build --root-owner-group pkg/swaypanion

clean:
	rm -fr swaypanion swaypanionc pkg