package project_test

import (
	"evo/internal/project"
	"evo/internal/test_helpers"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_DiscoverWorkspaces(t *testing.T) {
	var clearFixture, _, tmpDirAbs = test_helpers.RestoreFixture("simple_project")
	defer clearFixture()
	var workspaces = project.DiscoverWorkspaces(tmpDirAbs, []string{"**"})
	var workspacesList = []string{}
	workspaces.Range(func(key, value any) bool {
		workspacesList = append(workspacesList, key.(string))
		return true
	})
	assert.Equal(t, 3, len(workspacesList))
}

func Test_DiscoverWorkspacesWithNested(t *testing.T) {
	var clearFixture, _, tmpDirAbs = test_helpers.RestoreFixture("project_with_nested")
	defer clearFixture()
	var workspaces = project.DiscoverWorkspaces(tmpDirAbs, []string{"**", "packages/**", "packages/pkg-d/**"})
	var workspacesList = []string{}
	workspaces.Range(func(key, value any) bool {
		workspacesList = append(workspacesList, key.(string))
		return true
	})
	assert.Equal(t, 5, len(workspacesList))
}
