# Swaypanion: a companion app for Sway

Swaypanion combines different useful features for sway and similar desktop environments.

## Features

- brightness setting
<!-- - player control -->
- volume control
<!-- - dynamic workspaces -->
<!-- - selective window hiding -->
<!-- - _Freedesktop_ notifications -->

## Swaypanion daemon

Start the *Swaypanion* daemon with the `swaypanion` command.

If you have installed *Swaypanion* as a package, you may use the systemd service:

```plain
systemctl --user start swaypanion
```

The following command ensures *Swaypanion* is automatically started when you log in:

```plain
systemctl --user enable swaypanion
```

## Swaypanion client

The `swaypanionc` interactive command-line tool may be used to send commands and receive responses. Use the `help` command in the `swaypanionc` command-line shell to get the list of all available commands.

Use "`:`" as a separator between a command and its argument(s). For instance:

```plain
> volume set:50
```

You may provide partial command names, as long as there is no ambiguity. For instance, `b d` translates to `brightness down`, however `b u` could mean `brightness up` or `brightness unsubscribe`: in this situation, you may use `b up`.

### One-shot use

If you provide commands as arguments to the `swaypanionc` command, these commands are executed, without the integrated shell.

For instance:

```plain
$ swaypanionc "volume up" "brightness down"
$ 
```

## Socket

The*Swaypanion* daemon listens for connections on a UNIX socket, usually in `/run/user/<pid>/swaypanion.sock`. Its protocol is simple: when a client connects, *Swaypanion* waits for a command and its arguments. The arguments separator is the _group separator_ (character `0x1d`), the end of command is the _end of record_ character (`0x1e`).

## Notifications
