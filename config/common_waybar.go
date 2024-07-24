package config

import (
	"strconv"

	"github.com/willoma/swaypanion/common"
)

type WaybarPercent struct {
	Icon0          string   `yaml:"icon0"`
	Icons          []string `yaml:"icons"`
	TextFormat0    string   `yaml:"text_format0"`
	TextFormats    []string `yaml:"text_formats"`
	TooltipFormat0 string   `yaml:"tooltip_format0"`
	TooltipFormats []string `yaml:"tooltip_formats"`
}

func (w WaybarPercent) applyDefault(def WaybarPercent) WaybarPercent {
	if w.Icon0 == "" {
		w.Icon0 = def.Icon0
	}

	if len(w.Icons) == 0 {
		w.Icons = make([]string, len(def.Icons))
		copy(w.Icons, def.Icons)
	}

	if w.TextFormat0 == "" {
		w.TextFormat0 = def.TextFormat0
	}

	if len(w.TextFormats) == 0 {
		w.TextFormats = make([]string, len(def.TextFormats))
		copy(w.TextFormats, def.TextFormats)
	}

	if w.TooltipFormat0 == "" {
		w.TooltipFormat0 = def.TooltipFormat0
	}

	if len(w.TooltipFormats) == 0 {
		w.TooltipFormats = make([]string, len(def.TooltipFormats))
		copy(w.TooltipFormats, def.TooltipFormats)
	}

	return w
}

func (w WaybarPercent) FormatValue(valueStr string) (icon, text, tooltip string, ok bool) {
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		common.LogError(`Failed to convert "`+valueStr+`" to int`, err)
		return "", "", "", false
	}

	var (
		textFormat    string
		tooltipFormat string
	)

	if value <= 0 {
		icon = w.Icon0
		textFormat = w.TextFormat0
		tooltipFormat = w.TooltipFormat0
	} else if value >= 100 {
		icon = w.Icons[len(w.Icons)-1]
		textFormat = w.TextFormats[len(w.TextFormats)-1]
		tooltipFormat = w.TooltipFormats[len(w.TooltipFormats)-1]
	} else {
		stepSize := 100 / len(w.Icons)
		icon = w.Icons[value/stepSize]
		stepSize = 100 / len(w.TextFormats)
		textFormat = w.TextFormats[value/stepSize]
		stepSize = 100 / len(w.TooltipFormats)
		tooltipFormat = w.TooltipFormats[value/stepSize]
	}

	text = common.ReplaceValue(textFormat, "value", valueStr)
	tooltip = common.ReplaceValue(tooltipFormat, "value", valueStr)

	return icon, text, tooltip, true
}

type WaybarString struct {
	Icons         map[string]string `yaml:"icons"`
	FormatText    string            `yaml:"format_text"`
	FormatTooltip string            `yaml:"format_tooltip"`
}

func (w WaybarString) applyDefault(def WaybarString) WaybarString {
	if len(w.Icons) == 0 {
		w.Icons = make(map[string]string, len(def.Icons))
		for k, v := range def.Icons {
			w.Icons[k] = v
		}
	}
	if w.FormatText == "" {
		w.FormatText = def.FormatText
	}

	if w.FormatTooltip == "" {
		w.FormatTooltip = def.FormatTooltip
	}

	return w
}

func (w WaybarString) FormatValue(iconKey string, replaces map[string]string) (icon, text, tooltip string) {
	icon = w.Icons[iconKey]
	text = common.ReplaceValue(w.FormatText, "icon", icon)
	tooltip = common.ReplaceValue(w.FormatTooltip, "icon", icon)

	for name, content := range replaces {
		text = common.ReplaceValue(text, name, content)
		tooltip = common.ReplaceValue(tooltip, name, content)
	}

	return icon, text, tooltip
}
