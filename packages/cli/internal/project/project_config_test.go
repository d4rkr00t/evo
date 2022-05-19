package project_test

import (
	"evo/internal/project"
	"evo/internal/test_helpers"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_FindProjectConfigTopLevel(t *testing.T) {
	var clearFixture, _, tmpDirAbs = test_helpers.RestoreFixture("project")
	defer clearFixture()
	var _, err = project.FindProjectConfig(tmpDirAbs)
	assert.NoError(t, err)
}

func Test_FindProjectConfigFromNestedFolder(t *testing.T) {
	var clearFixture, _, tmpDirAbs = test_helpers.RestoreFixture("project")
	defer clearFixture()
	var nestedTmpDirAbs = path.Join(tmpDirAbs, "nested", "folder")
	var rootCfgPath, _ = project.FindProjectConfig(tmpDirAbs)
	var nestedCfgPath, err = project.FindProjectConfig(nestedTmpDirAbs)
	assert.NoError(t, err)
	assert.Equal(t, rootCfgPath, nestedCfgPath)
}

func Test_LoadConfig(t *testing.T) {
	var clearFixture, _, tmpDirAbs = test_helpers.RestoreFixture("project")
	defer clearFixture()
	var cfgPath, _ = project.FindProjectConfig(tmpDirAbs)
	var _, err = project.LoadConfig(cfgPath)
	assert.NoError(t, err)
}

func Test_GetExcludesForWorkspace(t *testing.T) {
	var clearFixture, _, tmpDirAbs = test_helpers.RestoreFixture("project")
	defer clearFixture()
	var cfgPath, _ = project.FindProjectConfig(tmpDirAbs)
	var wsPath = path.Join(tmpDirAbs, "nested", "folder")
	var cfg, _ = project.LoadConfig(cfgPath)
	var excludes = cfg.GetExcludes(tmpDirAbs, wsPath)
	assert.Equal(t, 2, len(excludes))
}
