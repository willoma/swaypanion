package swaypanion

import (
	"fmt"
	"io"
)

const (
	defaultWorkspacesEmptyWorkspaceName = " "
)

type WorkspacesConfig struct {
	Disable            bool   `yaml:"disable"`
	EmptyWorkspaceName string `yaml:"empty_workspace_name"`
}

func (c WorkspacesConfig) withDefaults() WorkspacesConfig {
	if c.EmptyWorkspaceName == "" {
		c.EmptyWorkspaceName = defaultWorkspacesEmptyWorkspaceName
	}

	return c
}

type Workspaces struct {
	conf WorkspacesConfig
	sway *SwayClient
}

func NewWorkspaces(conf WorkspacesConfig, sway *SwayClient) (*Workspaces, error) {
	if conf.Disable {
		return nil, nil
	}

	return &Workspaces{conf, sway}, nil
}

func (a *Workspaces) GoToPrevious() error {
	rank, workspaces, err := currentWorkspace(a.sway)
	if err != nil {
		return err
	}

	if rank != 0 {
		// Previous workspace exists, simply go!
		return a.sway.RunCommand("workspace prev_on_output")
	}

	if nbWindows := countWindows(workspaces[0]); nbWindows == 0 {
		// No windows in current workspace, stay here
		return nil
	}

	// Short path: the workspace has a number larger than 1
	if num, _ := extractWorkspaceName(workspaces[0]); num > 1 {
		return a.sway.RunCommand(fmt.Sprintf(
			`workspace "%s"`, makeWorkspaceName(num-1, a.conf.EmptyWorkspaceName),
		))
	}

	// Long path: rename all workspaces...
	if err := a.sway.renameWorkspacesInOrder(workspaces); err != nil {
		return err
	}

	return a.sway.RunCommand(fmt.Sprintf(
		`workspace "%s"`, makeWorkspaceName(1, a.conf.EmptyWorkspaceName),
	))
}

func (a *Workspaces) GoToNext() error {
	rank, workspaces, err := currentWorkspace(a.sway)
	if err != nil {
		return err
	}

	if rank != len(workspaces)-1 {
		// Next workspace exists, simply go!
		return a.sway.RunCommand("workspace next_on_output")
	}

	if nbWindows := countWindows(workspaces[rank]); nbWindows == 0 {
		// No windows in current workspace, stay here
		return nil
	}

	// Short path: the workspace has a number
	if num, _ := extractWorkspaceName(workspaces[rank]); num != -1 {
		return a.sway.RunCommand(fmt.Sprintf(
			`workspace "%s"`, makeWorkspaceName(num+1, a.conf.EmptyWorkspaceName),
		))
	}

	// Long path: rename all workspaces...
	if err := a.sway.renameWorkspacesInOrder(workspaces); err != nil {
		return err
	}

	return a.sway.RunCommand(fmt.Sprintf(
		`workspace "%s"`, makeWorkspaceName(rank+1, a.conf.EmptyWorkspaceName),
	))
}

func (a *Workspaces) MoveToPrevious() error {
	rank, workspaces, err := currentWorkspace(a.sway)
	if err != nil {
		return err
	}

	if rank != 0 {
		// Previous workspace exists, simply go!
		return a.sway.RunCommand("[con_id=__focused__] move to workspace prev_on_output, focus")
	}

	if nbWindows := countWindows(workspaces[0]); nbWindows <= 1 {
		// Only one window in current workspace, stay here
		return nil
	}

	// Short path: the workspace has a number larger than 1
	if num, _ := extractWorkspaceName(workspaces[0]); num > 1 {
		return a.sway.RunCommand(fmt.Sprintf(
			`[con_id=__focused__] move to workspace "%s", focus`,
			makeWorkspaceName(num-1, a.conf.EmptyWorkspaceName),
		))
	}

	// Long path: rename all workspaces...
	if err := a.sway.renameWorkspacesInOrder(workspaces); err != nil {
		return err
	}

	return a.sway.RunCommand(fmt.Sprintf(
		`[con_id=__focused__] move to workspace "%s", focus`, makeWorkspaceName(1, a.conf.EmptyWorkspaceName),
	))
}

func (a *Workspaces) MoveToNext() error {
	rank, workspaces, err := currentWorkspace(a.sway)
	if err != nil {
		return err
	}

	if rank != len(workspaces)-1 {
		// Next workspace exists, simply go!
		return a.sway.RunCommand("[con_id=__focused__] move to workspace next_on_output, focus")
	}

	if nbWindows := countWindows(workspaces[rank]); nbWindows <= 1 {
		// Only one window in current workspace, stay here
		return nil
	}

	// Short path: the workspace has a number
	if num, _ := extractWorkspaceName(workspaces[rank]); num != -1 {
		return a.sway.RunCommand(fmt.Sprintf(
			`[con_id=__focused__] move to workspace "%s", focus`, makeWorkspaceName(num+1, a.conf.EmptyWorkspaceName),
		))
	}

	// Long path: rename all workspaces...
	if err := a.sway.renameWorkspacesInOrder(workspaces); err != nil {
		return err
	}

	return a.sway.RunCommand(fmt.Sprintf(
		`[con_id=__focused__] move to workspace "%s", focus`, makeWorkspaceName(rank+1, a.conf.EmptyWorkspaceName),
	))
}

func (s *Swaypanion) dynworkspaceHandler(args []string, w io.Writer) error {
	if len(args) == 0 {
		return missingArgument("action")
	}

	switch args[0] {
	case "previous":
		return s.workspaces.GoToPrevious()
	case "next":
		return s.workspaces.GoToNext()
	case "move":
		if len(args) < 2 {
			return missingArgument("movement target")
		}

		switch args[1] {
		case "previous":
			return s.workspaces.MoveToPrevious()
		case "next":
			return s.workspaces.MoveToNext()
		default:
			return s.unknownArgument(w, "dynworkspace", args[1])

		}
	default:
		return s.unknownArgument(w, "dynworkspace", args[0])
	}
}

func (s *Swaypanion) registerWorkspaces() {
	if s.conf.Workspaces.Disable {
		return
	}

	s.register(
		"dynworkspace", s.dynworkspaceHandler, "move to workspace with dynamic creation if needed",
		"previous", "go to the previous workspace, creating it if needed",
		"next", "go to the next workspace, creating it if needed",
		"move previous", "move the focused window to the previous workspace, creating it if needed",
		"move next", "move the focused window to the next workspace, creating it if needed",
	)
}
