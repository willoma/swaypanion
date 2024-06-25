# Swaypanion: a companion app for Sway

Swaypanion tries to combine different useful features for sway and similar desktop environments.

## Features

- brightness setting, with _Freedesktop_ notification
- player control
- volume control, with _Freedesktop_ notification
- dynamic workspaces
- selective window hiding

## Socket and client

*Swaypanion* listens on to a UNIX socket usually in `/run/user/<pid>/swaypanion.sock`. Its protocol is simple: when a client connects, *Swaypanion* waits for a command and its arguments. The arguments separator is the _unit separator_ (character `0x1f`), the end of command is the _enquiry_ character (`0x05`).

The `swaypanionc` command may be used to send commands and receive responses.

The `help` command lists all available commands.
