package config

type SwayNodes struct {
	config[*SwayNodes] `yaml:"-"`

	EmptyWorkspaceName string                 `yaml:"empty_workspace_name"`
	HideInsteadOfClose []WindowIdentification `yaml:"hide_instead_of_close"`
}

var DefaultSwayNodes = &SwayNodes{
	EmptyWorkspaceName: " ",
	HideInsteadOfClose: []WindowIdentification{
		{
			Type:    WindowMatchInstance,
			Partial: false,
			Match:   "spotify",
		},
	},
}

func (s *SwayNodes) applyDefault() {
	if s.EmptyWorkspaceName == "" {
		s.EmptyWorkspaceName = DefaultSwayNodes.EmptyWorkspaceName
	}

	if len(s.HideInsteadOfClose) == 0 {
		s.HideInsteadOfClose = make([]WindowIdentification, len(DefaultSwayNodes.HideInsteadOfClose))
		copy(s.HideInsteadOfClose, DefaultSwayNodes.HideInsteadOfClose)
	}
}
