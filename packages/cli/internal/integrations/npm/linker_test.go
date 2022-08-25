package npm_test

import (
	"evo/internal/integrations/npm"
	"evo/internal/project"
	"evo/internal/test_helpers"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Npm_Linker_Local(t *testing.T) {
	var clearFixture, _, tmpDirAbs = test_helpers.RestoreFixture("simple")
	defer clearFixture()

	var proj, _ = project.NewProject(path.Join(tmpDirAbs, project.ProjectConfigFileName))

	for _, wsName := range proj.WorkspacesNames {
		npm.AddNpmDependencies(&proj, wsName)
	}

	npm.LinkNpmDependencies(&proj, "pkg-b")
	assert.FileExists(t, path.Join(tmpDirAbs, "pkg-b", "node_modules", "pkg-c"))
	assert.FileExists(t, path.Join(tmpDirAbs, "pkg-b", "node_modules", ".bin", "pkgc"))
}

// func Test_Npm_Linker_External(t *testing.T) {
// 	var clearFixture, _, tmpDirAbs = test_helpers.RestoreFixture("external")
// 	defer clearFixture()

// 	var proj, _ = project.NewProject(path.Join(tmpDirAbs, project.ProjectConfigFileName))

// 	for _, wsName := range proj.WorkspacesNames {
// 		npm.AddNpmDependencies(&proj, wsName)
// 	}

// 	npm.LinkNpmDependencies(&proj, "pkg-a")

// 	assert.FileExists(t, path.Join(tmpDirAbs, "packages", "pkg-a", "node_modules", "typescript"))
// 	assert.FileExists(t, path.Join(tmpDirAbs, "packages", "pkg-a", "node_modules", ".bin", "tsc"))
// }

// func Test_Npm_Linker_External_String_Bin(t *testing.T) {
// 	var _, _, tmpDirAbs = test_helpers.RestoreFixture("external_string_bin")
// 	// defer clearFixture()

// 	var proj, _ = project.NewProject(path.Join(tmpDirAbs, project.ProjectConfigFileName))

// 	for _, wsName := range proj.WorkspacesNames {
// 		npm.AddNpmDependencies(&proj, wsName)
// 	}

// 	npm.LinkNpmDependencies(&proj, "pkg-a")
// 	assert.FileExists(t, path.Join(tmpDirAbs, "packages", "pkg-a", "node_modules", "typescript"))
// 	assert.FileExists(t, path.Join(tmpDirAbs, "packages", "pkg-a", "node_modules", ".bin", "typescript"))
// }
