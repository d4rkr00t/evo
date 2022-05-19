package npm_test

import (
	"evo/internal/integrations/npm"
	"evo/internal/project"
	"evo/internal/test_helpers"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Npm_Dependencies(t *testing.T) {
	var clearFixture, _, tmpDirAbs = test_helpers.RestoreFixture("simple")
	defer clearFixture()

	var proj, _ = project.NewProject(path.Join(tmpDirAbs, project.ProjectConfigFileName))

	for _, wsName := range proj.WorkspacesNames {
		npm.AddNpmDependencies(&proj, wsName)
	}

	var ws, _ = proj.Load("pkg-a")
	assert.True(t, len(ws.Deps) > 0)
}
