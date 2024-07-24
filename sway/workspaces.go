package sway

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/joshuarubin/go-sway"
)

var (
	ErrCurrentWorkspaceNotFound = errors.New("current workspace not found")

	reWorkspaceName = regexp.MustCompile("([0-9]+): (.*)")
)

func (c *Client) CurrentWorkspace() (rankInOutput int, workspaces []*sway.Node, err error) {
	root, err := c.client.GetTree(c.ctx)
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

func (c *Client) RenameWorkspacesInOrder(workspaces []*sway.Node) error {
	for i := len(workspaces) - 1; i >= 0; i-- {
		_, name := ExtractWorkspaceName(workspaces[i])
		if err := c.RunCommand(fmt.Sprintf(
			`rename workspace %q to %q"`,
			workspaces[i].Name, MakeWorkspaceName(i+1, name),
		)); err != nil {
			return err
		}
	}

	return nil
}

func ExtractWorkspaceName(workspace *sway.Node) (num int, name string) {
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

func MakeWorkspaceName(num int, name string) string {
	return strconv.Itoa(num) + ": " + strings.ReplaceAll(name, `"`, `\"`)
}
