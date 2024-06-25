package swaypanion

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/joshuarubin/go-sway"
)

var (
	ErrCurrentWindowNotFound    = errors.New("current window not found")
	ErrCurrentWorkspaceNotFound = errors.New("current workspace not found")

	reWorkspaceName = regexp.MustCompile("([0-9]+): (.*)")
)

type SwayClient struct {
	client sway.Client

	ctx   context.Context
	Close context.CancelFunc
}

func NewSwayClient() (*SwayClient, error) {
	ctx, cancel := context.WithCancel(context.Background())

	client, err := sway.New(ctx)
	if err != nil {
		cancel()
		return nil, err
	}

	return &SwayClient{
		client: client,
		ctx:    ctx,
		Close:  cancel,
	}, nil
}

func (s *SwayClient) GetTree() (*sway.Node, error) {
	return s.client.GetTree(s.ctx)
}

func (s *SwayClient) RunCommand(command string) error {
	resp, err := s.client.RunCommand(s.ctx, command)
	if err != nil {
		return err
	}

	for _, r := range resp {
		if !r.Success {
			return errors.New(r.Error)
		}
	}

	return nil
}

func (client *SwayClient) focusedNode() (*sway.Node, error) {
	root, err := client.GetTree()
	if err != nil {
		return nil, err
	}

	focused := root.FocusedNode()

	if focused == nil {
		return nil, ErrCurrentWindowNotFound
	}

	return focused, nil

}

func (client *SwayClient) currentWorkspace() (rankInOutput int, workspaces []*sway.Node, err error) {
	root, err := client.GetTree()
	if err != nil {
		return 0, nil, err
	}

	for _, output := range root.Nodes {
		if output.Type != sway.NodeOutput {
			continue
		}

		var (
			focusedWorkspaceRank = -1
			workspaces           []*sway.Node
		)

		for i, workspace := range output.Nodes {
			if workspace.Type != sway.NodeWorkspace {
				continue
			}

			workspaces = append(workspaces, workspace)

			if hasFocus(workspace) {
				focusedWorkspaceRank = i
			}
		}

		if focusedWorkspaceRank != -1 {
			return focusedWorkspaceRank, workspaces, nil
		}
	}

	return 0, nil, ErrCurrentWorkspaceNotFound
}

func (s *SwayClient) renameWorkspacesInOrder(workspaces []*sway.Node) error {
	for i := len(workspaces) - 1; i >= 0; i-- {
		_, name := extractWorkspaceName(workspaces[i])
		if err := s.RunCommand(fmt.Sprintf(
			`rename workspace %q to %q"`,
			workspaces[i].Name, makeWorkspaceName(i+1, name),
		)); err != nil {
			return err
		}
	}

	return nil
}

func extractWorkspaceName(workspace *sway.Node) (num int, name string) {
	matches := reWorkspaceName.FindStringSubmatch(workspace.Name)
	if matches == nil {
		return -1, name
	}

	num, err := strconv.Atoi(matches[1])
	if err != nil {
		num = -1
	}

	return num, matches[2]
}

func makeWorkspaceName(num int, name string) string {
	return strconv.Itoa(num) + ": " + strings.ReplaceAll(name, `"`, `\"`)
}

func countWindows(node *sway.Node) int {
	if len(node.Nodes) == 0 && len(node.FloatingNodes) == 0 {
		return 1
	}

	var count int

	for _, subnode := range node.Nodes {
		count += countWindows(subnode)
	}

	for _, subnode := range node.FloatingNodes {
		count += countWindows(subnode)
	}

	return count
}

func hasFocus(node *sway.Node) bool {
	if node.Focused {
		return true
	}

	for _, subnode := range node.Nodes {
		if hasFocus(subnode) {
			return true
		}
	}

	for _, subnode := range node.FloatingNodes {
		if hasFocus(subnode) {
			return true
		}
	}

	return false
}
