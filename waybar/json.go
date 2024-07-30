package waybar

import (
	"bytes"
	"io"
	"strings"
)

var jsonReplacer = strings.NewReplacer(
	"\n", "\r",
	"\\r", "\r",
	"\\n", "\r",
	`"`, `\"`,
)

func writeJSON(w io.Writer, alt, text, tooltip string, disabled bool) {
	b := &bytes.Buffer{}

	b.WriteByte('{')

	var hasPrevious bool

	if alt != "" {
		b.WriteString(`"alt":"`)
		jsonReplacer.WriteString(b, alt)
		b.WriteByte('"')
		hasPrevious = true
	}

	if text != "" {
		if hasPrevious {
			b.WriteByte(',')
		}
		b.WriteString(`"text":"`)
		jsonReplacer.WriteString(b, text)
		b.WriteByte('"')
		hasPrevious = true
	}

	if tooltip != "" {
		if hasPrevious {
			b.WriteByte(',')
		}
		b.WriteString(`"tooltip":"`)
		jsonReplacer.WriteString(b, tooltip)
		b.WriteByte('"')
		hasPrevious = true
	}

	if disabled {
		if hasPrevious {
			b.WriteByte(',')
		}
		b.WriteString(`"class":"disabled"`)
	}

	b.WriteString("}\n")

	b.WriteTo(w)
}
