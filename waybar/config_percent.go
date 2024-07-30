package waybar

import (
	"strconv"

	"github.com/willoma/swaypanion/common"
)

type configPercent struct {
	IconDisabled          string   `yaml:"icon_disabled"`
	Icon0                 string   `yaml:"icon0"`
	Icons                 []string `yaml:"icons"`
	TextFormatDisabled    string   `yaml:"text_format_disabled"`
	TextFormat0           string   `yaml:"text_format0"`
	TextFormats           []string `yaml:"text_formats"`
	TooltipFormatDisabled string   `yaml:"tooltip_format_disabled"`
	TooltipFormat0        string   `yaml:"tooltip_format0"`
	TooltipFormats        []string `yaml:"tooltip_formats"`
}

func (c configPercent) applyDefault(def configPercent) configPercent {
	if c.IconDisabled == "" {
		c.IconDisabled = def.IconDisabled
	}

	if c.Icon0 == "" {
		c.Icon0 = def.Icon0
	}

	if len(c.Icons) == 0 {
		c.Icons = make([]string, len(def.Icons))
		copy(c.Icons, def.Icons)
	}

	if c.TextFormatDisabled == "" {
		c.TextFormatDisabled = def.TextFormatDisabled
	}

	if c.TextFormat0 == "" {
		c.TextFormat0 = def.TextFormat0
	}

	if len(c.TextFormats) == 0 {
		c.TextFormats = make([]string, len(def.TextFormats))
		copy(c.TextFormats, def.TextFormats)
	}

	if c.TooltipFormatDisabled == "" {
		c.TooltipFormatDisabled = def.TooltipFormatDisabled
	}

	if c.TooltipFormat0 == "" {
		c.TooltipFormat0 = def.TooltipFormat0
	}

	if len(c.TooltipFormats) == 0 {
		c.TooltipFormats = make([]string, len(def.TooltipFormats))
		copy(c.TooltipFormats, def.TooltipFormats)
	}

	return c
}

func (c configPercent) formatValue(valueStr string) (icon, text, tooltip string, disabled bool) {
	value, err := strconv.Atoi(valueStr)
	disabled = err != nil

	var (
		textFormat    string
		tooltipFormat string
	)

	if disabled {
		icon = c.IconDisabled
		textFormat = c.TextFormatDisabled
		tooltipFormat = c.TooltipFormatDisabled
	} else if value <= 0 {
		icon = c.Icon0
		textFormat = c.TextFormat0
		tooltipFormat = c.TooltipFormat0
	} else if value >= 100 {
		icon = c.Icons[len(c.Icons)-1]
		textFormat = c.TextFormats[len(c.TextFormats)-1]
		tooltipFormat = c.TooltipFormats[len(c.TooltipFormats)-1]
	} else {
		stepSize := 100 / len(c.Icons)
		icon = c.Icons[value/stepSize]
		stepSize = 100 / len(c.TextFormats)
		textFormat = c.TextFormats[value/stepSize]
		stepSize = 100 / len(c.TooltipFormats)
		tooltipFormat = c.TooltipFormats[value/stepSize]
	}

	text = common.ReplaceValue(textFormat, "value", valueStr)
	tooltip = common.ReplaceValue(tooltipFormat, "value", valueStr)

	return icon, text, tooltip, disabled
}
