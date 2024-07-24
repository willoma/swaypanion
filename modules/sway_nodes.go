package modules

import (
	"fmt"
	"sync"

	"github.com/willoma/swaypanion/common"
	"github.com/willoma/swaypanion/config"
	"github.com/willoma/swaypanion/sway"
)

type SwayNodes struct {
	sway *sway.Client
	stop func()

	mu                 sync.Mutex
	emptyWorkspaceName string
	hideInsteadOfClose []config.WindowIdentification
}

func NewSwayNodes(conf *config.SwayNodes, swayClient *sway.Client) *SwayNodes {
	s := &SwayNodes{
		sway: swayClient,
	}

	s.reloadConfig(conf)
	s.stop = conf.ListenReload(s.reloadConfig)

	return s
}

func (s *SwayNodes) reloadConfig(conf *config.SwayNodes) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.emptyWorkspaceName = conf.EmptyWorkspaceName
	s.hideInsteadOfClose = make([]config.WindowIdentification, len(conf.HideInsteadOfClose))
	copy(s.hideInsteadOfClose, conf.HideInsteadOfClose)
}

func (s *SwayNodes) moveNext() {
	rank, workspaces, err := s.sway.CurrentWorkspace()
	if err != nil {
		common.LogError("Failed to get current workspace", err)
		return
	}

	if rank != len(workspaces)-1 {
		// Next workspace exists, simply go!
		if err := s.sway.RunCommand("[con_id=__focused__] move to workspace next_on_output, focus"); err != nil {
			common.LogError("Failed to move to next workspace", err)
			return
		}
	}

	if nbWindows := sway.CountWindows(workspaces[rank]); nbWindows <= 1 {
		// Only one window in current workspace, stay here
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Short path: the workspace has a number
	if num, _ := sway.ExtractWorkspaceName(workspaces[rank]); num != -1 {
		if err := s.sway.RunCommand(fmt.Sprintf(
			`[con_id=__focused__] move to workspace "%s", focus`, sway.MakeWorkspaceName(num+1, s.emptyWorkspaceName),
		)); err != nil {
			common.LogError("Failed to move to newly created next workspace", err)
			return
		}
	}

	// Long path: rename all workspaces...
	if err := s.sway.RenameWorkspacesInOrder(workspaces); err != nil {
		common.LogError("Failed to rename workspaces in order", err)
		return
	}

	if err := s.sway.RunCommand(fmt.Sprintf(
		`[con_id=__focused__] move to workspace "%s", focus`, sway.MakeWorkspaceName(rank+1, s.emptyWorkspaceName),
	)); err != nil {
		common.LogError("Failed to move to newly created next workspace", err)
	}
}

func (s *SwayNodes) movePrev() {
	rank, workspaces, err := s.sway.CurrentWorkspace()
	if err != nil {
		common.LogError("Failed to get current workspace", err)
		return
	}

	if rank != 0 {
		// Previous workspace exists, simply go!
		if err := s.sway.RunCommand("[con_id=__focused__] move to workspace prev_on_output, focus"); err != nil {
			common.LogError("Failed to move to previous workspace", err)
			return
		}
	}

	if nbWindows := sway.CountWindows(workspaces[0]); nbWindows <= 1 {
		// Only one window in current workspace, stay here
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Short path: the workspace has a number larger than 1
	if num, _ := sway.ExtractWorkspaceName(workspaces[0]); num > 1 {
		if err := s.sway.RunCommand(fmt.Sprintf(
			`[con_id=__focused__] move to workspace "%s", focus`,
			sway.MakeWorkspaceName(num-1, s.emptyWorkspaceName),
		)); err != nil {
			common.LogError("Failed to move to newly created previous workspace", err)
			return
		}
	}

	// Long path: rename all workspaces...
	if err := s.sway.RenameWorkspacesInOrder(workspaces); err != nil {
		common.LogError("Failed to rename workspaces in order", err)
		return
	}

	if err := s.sway.RunCommand(fmt.Sprintf(
		`[con_id=__focused__] move to workspace "%s", focus`, sway.MakeWorkspaceName(1, s.emptyWorkspaceName),
	)); err != nil {
		common.LogError("Failed to move to newly created previous workspace", err)
	}
}

func (s *SwayNodes) next() {
	rank, workspaces, err := s.sway.CurrentWorkspace()
	if err != nil {
		common.LogError("Failed to get current workspace", err)
		return
	}

	if rank != len(workspaces)-1 {
		// Next workspace exists, simply go!
		if err := s.sway.RunCommand("workspace next_on_output"); err != nil {
			common.LogError("Failed to go to next workspace", err)
			return
		}
	}

	if nbWindows := sway.CountWindows(workspaces[rank]); nbWindows == 0 {
		// No windows in current workspace, stay here
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Short path: the workspace has a number
	if num, _ := sway.ExtractWorkspaceName(workspaces[rank]); num != -1 {
		if err := s.sway.RunCommand(fmt.Sprintf(
			`workspace "%s"`, sway.MakeWorkspaceName(num+1, s.emptyWorkspaceName),
		)); err != nil {
			common.LogError("Failed to go to newly created next workspace", err)
			return
		}
	}

	// Long path: rename all workspaces...
	if err := s.sway.RenameWorkspacesInOrder(workspaces); err != nil {
		common.LogError("Failed to rename workspaces in order", err)
		return
	}

	if err := s.sway.RunCommand(fmt.Sprintf(
		`workspace "%s"`, sway.MakeWorkspaceName(rank+1, s.emptyWorkspaceName),
	)); err != nil {
		common.LogError("Failed to go to newly created next workspace", err)
	}
}

func (s *SwayNodes) prev() {
	rank, workspaces, err := s.sway.CurrentWorkspace()
	if err != nil {
		common.LogError("Failed to get current workspace", err)
		return
	}

	if rank != 0 {
		// Previous workspace exists, simply go!
		if err := s.sway.RunCommand("workspace prev_on_output"); err != nil {
			common.LogError("Failed to go to previous workspace", err)
			return
		}
	}

	if nbWindows := sway.CountWindows(workspaces[0]); nbWindows == 0 {
		// No windows in current workspace, stay here
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Short path: the workspace has a number larger than 1
	if num, _ := sway.ExtractWorkspaceName(workspaces[0]); num > 1 {
		if err := s.sway.RunCommand(fmt.Sprintf(
			`workspace "%s"`, sway.MakeWorkspaceName(num-1, s.emptyWorkspaceName),
		)); err != nil {
			common.LogError("Failed to go to newly created previous workspace", err)
			return
		}
	}

	// Long path: rename all workspaces...
	if err := s.sway.RenameWorkspacesInOrder(workspaces); err != nil {
		common.LogError("Failed to rename workspaces in order", err)
		return
	}

	if err := s.sway.RunCommand(fmt.Sprintf(
		`workspace "%s"`, sway.MakeWorkspaceName(1, s.emptyWorkspaceName),
	)); err != nil {
		common.LogError("Failed to go to newly created previous workspace", err)
	}
}

func (s *SwayNodes) hideOrClose() {
	window, err := s.sway.FocusedNode()
	if err != nil {
		common.LogError("Failed to get currently focused window", err)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, id := range s.hideInsteadOfClose {
		if id.MatchWindow(window) {
			if err := s.sway.RunCommand("move to scratchpad"); err != nil {
				common.LogError("Failed to hide window", err)
			}
			return
		}
	}

	if err := s.sway.RunCommand("kill"); err != nil {
		common.LogError("Failed to close window", err)
	}
}
