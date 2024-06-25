package swaypanion

import "github.com/godbus/dbus/v5"

func dbusExtractPlayerMetadata(data any) playerMetadata {
	result := playerMetadata{}

	if array, ok := data.(map[string]dbus.Variant); ok {
		for k, v := range array {
			switch k {
			case "xesam:album":
				result.album = v.Value().(string)
			case "xesam:artist":
				result.artists = v.Value().([]string)
			case "xesam:title":
				result.title = v.Value().(string)
			}
		}
	}

	return result
}
