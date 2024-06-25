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

## Commands

| Command | Result |
|-|-|
| **Brightness** | |
| `brightness` | show the screen brightness in percent |
| `brightness up` | make the screen brighter |
| `brightness down` | make the screen dimmer |
| `brightness set X` | set the screen brightness to X percent |
| **Player** | |
| `player` | start or show the audio player |
| `player playpause` | play or pause the music |
| `player previous` | go back to previous track |
| `player next` | skip to next track |
| **Volume** | |
| `volume` | show the volume in percent |
| `volume up` | make the volume higher |
| `volume down` | make the volume lower |
| `volume mute` | toggle the mute status |
| `volume set X` | set the volume to X percent |
| **Windows** | |
| `window hide_or_close` | hide or kill the focused window, according to the hide_not_close configuration |
| **Workspaces** | |
| `dynworkspace previous` | go to the previous workspace, creating it if needed |
| `dynworkspace next` | go to the next workspace, creating it if needed |
| `dynworkspace move previous` | move the focused window to the previous workspace, creating it if needed |
| `dynworkspace move next` | move the focused window to the next workspace, creating it if needed |

## Notifications

- Level notification on volume change
