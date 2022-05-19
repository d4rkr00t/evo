package project_test

import (
	"evo/internal/project"
	"evo/internal/test_helpers"
	"evo/internal/workspace"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ProjectCreation(t *testing.T) {
	var clearFixture, _, tmpDirAbs = test_helpers.RestoreFixture("project")
	defer clearFixture()
	var proj, err = project.NewProject(path.Join(tmpDirAbs, project.ProjectConfigFileName))
	assert.NoError(t, err)
	assert.Equal(t, 1, proj.Size())
}

func Test_ProjectCreationDuplicates(t *testing.T) {
	var clearFixture, _, tmpDirAbs = test_helpers.RestoreFixture("project_duplicates")
	defer clearFixture()
	var _, err = project.NewProject(path.Join(tmpDirAbs, project.ProjectConfigFileName))
	assert.Error(t, err)
}

func Test_ProjectDependencyGraph(t *testing.T) {
	var clearFixture, _, tmpDirAbs = test_helpers.RestoreFixture("simple_project")
	defer clearFixture()
	var proj, _ = project.NewProject(path.Join(tmpDirAbs, project.ProjectConfigFileName))

	var ws, _ = proj.Load("pkg-a")
	ws.Deps["pkg-b"] = workspace.WorkspaceDependency{
		Name:     "pkg-b",
		Version:  "*",
		Provider: "test",
		Type:     "local",
	}

	ws, _ = proj.Load("pkg-b")
	ws.Deps["pkg-c"] = workspace.WorkspaceDependency{
		Name:     "pkg-c",
		Version:  "*",
		Provider: "test",
		Type:     "local",
	}

	proj.BuildDependencyGraph()

	var topologicalOrder = []string{}
	proj.Walk(func(ws *workspace.Workspace) error {
		topologicalOrder = append(topologicalOrder, ws.Name)
		return nil
	}, 3)

	assert.Equal(t, topologicalOrder[0], "pkg-c")
	assert.Equal(t, topologicalOrder[1], "pkg-b")
	assert.Equal(t, topologicalOrder[2], "pkg-a")
}

func Test_ProjectWithCycles(t *testing.T) {
	var clearFixture, _, tmpDirAbs = test_helpers.RestoreFixture("simple_project")
	defer clearFixture()
	var proj, _ = project.NewProject(path.Join(tmpDirAbs, project.ProjectConfigFileName))

	var ws, _ = proj.Load("pkg-a")
	ws.Deps["pkg-b"] = workspace.WorkspaceDependency{
		Name:     "pkg-b",
		Version:  "*",
		Provider: "test",
		Type:     "local",
	}

	ws, _ = proj.Load("pkg-b")
	ws.Deps["pkg-c"] = workspace.WorkspaceDependency{
		Name:     "pkg-c",
		Version:  "*",
		Provider: "test",
		Type:     "local",
	}

	ws, _ = proj.Load("pkg-c")
	ws.Deps["pkg-a"] = workspace.WorkspaceDependency{
		Name:     "pkg-a",
		Version:  "*",
		Provider: "test",
		Type:     "local",
	}

	proj.BuildDependencyGraph()

	var err = proj.Validate()

	assert.Error(t, err)
}

func Test_ProjectReduceToScope(t *testing.T) {
	var clearFixture, _, tmpDirAbs = test_helpers.RestoreFixture("simple_project_5")
	defer clearFixture()
	var proj, _ = project.NewProject(path.Join(tmpDirAbs, project.ProjectConfigFileName))

	var ws, _ = proj.Load("pkg-a")
	ws.Deps["pkg-b"] = workspace.WorkspaceDependency{
		Name:     "pkg-b",
		Version:  "*",
		Provider: "test",
		Type:     "local",
	}
	ws.Deps["pkg-c"] = workspace.WorkspaceDependency{
		Name:     "pkg-c",
		Version:  "*",
		Provider: "test",
		Type:     "local",
	}

	ws, _ = proj.Load("pkg-b")
	ws.Deps["pkg-d"] = workspace.WorkspaceDependency{
		Name:     "pkg-d",
		Version:  "*",
		Provider: "test",
		Type:     "local",
	}

	proj.ReduceToScope([]string{"pkg-a"})

	assert.Contains(t, proj.WorkspacesNames, "pkg-a")
	assert.Contains(t, proj.WorkspacesNames, "pkg-b")
	assert.Contains(t, proj.WorkspacesNames, "pkg-c")
	assert.Contains(t, proj.WorkspacesNames, "pkg-d")
	assert.NotContains(t, proj.WorkspacesNames, "pkg-e")
}
