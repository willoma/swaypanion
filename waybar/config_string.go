package waybar

import "github.com/willoma/swaypanion/common"

type configString struct {
	Icons         map[string]string `yaml:"icons"`
	FormatText    string            `yaml:"format_text"`
	FormatTooltip string            `yaml:"format_tooltip"`
}

func (c configString) applyDefault(def configString) configString {
	if len(c.Icons) == 0 {
		c.Icons = make(map[string]string, len(def.Icons))
		for k, v := range def.Icons {
			c.Icons[k] = v
		}
	}
	if c.FormatText == "" {
		c.FormatText = def.FormatText
	}

	if c.FormatTooltip == "" {
		c.FormatTooltip = def.FormatTooltip
	}

	return c
}

func (c configString) formatValue(iconKey string, replaces map[string]string) (icon, text, tooltip string, disabled bool) {
	icon = c.Icons[iconKey]
	text = common.ReplaceValue(c.FormatText, "icon", icon)
	tooltip = common.ReplaceValue(c.FormatTooltip, "icon", icon)

	for name, content := range replaces {
		text = common.ReplaceValue(text, name, content)
		tooltip = common.ReplaceValue(tooltip, name, content)
	}

	return icon, text, tooltip, false
}
