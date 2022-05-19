package runner_test

import (
	"evo/internal/cache"
	"evo/internal/context"
	"evo/internal/logger"
	"evo/internal/project"
	"evo/internal/runner"
	"evo/internal/stats"
	"evo/internal/test_helpers"
	"evo/internal/tracer"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_BuildSimple(t *testing.T) {
	var clearFixture, _, tmpDirAbs = test_helpers.RestoreFixture("simple")
	defer clearFixture()

	var projectConfigPath, _ = project.FindProjectConfig(tmpDirAbs)
	var cache = cache.New(tmpDirAbs, cache.DefaultCacheLocation)
	cache.Setup()

	var ctx = context.Context{
		Root:              tmpDirAbs,
		Cwd:               tmpDirAbs,
		ProjectConfigPath: projectConfigPath,
		Targets:           []string{"build"},
		Concurrency:       2,
		ChangedFiles:      []string{},
		ChangedOnly:       false,
		Logger:            logger.NewLogger(false, false),
		Stats:             stats.New(),
		Tracer:            tracer.New(),
		Cache:             cache,
		Scope:             []string{},
	}

	var runErr = runner.Run(&ctx)

	assert.NoError(t, runErr)
	assert.FileExists(t, path.Join(tmpDirAbs, "packages", "pkg-a", "dist.js"))
	assert.FileExists(t, path.Join(tmpDirAbs, "packages", "pkg-b", "dist.js"))
}

func Test_BuildSimpleUnknownTarget(t *testing.T) {
	var clearFixture, _, tmpDirAbs = test_helpers.RestoreFixture("simple")
	defer clearFixture()

	var projectConfigPath, _ = project.FindProjectConfig(tmpDirAbs)
	var cache = cache.New(tmpDirAbs, cache.DefaultCacheLocation)
	cache.Setup()

	var ctx = context.Context{
		Root:              tmpDirAbs,
		Cwd:               tmpDirAbs,
		ProjectConfigPath: projectConfigPath,
		Targets:           []string{"test"},
		Concurrency:       2,
		ChangedFiles:      []string{},
		ChangedOnly:       false,
		Logger:            logger.NewLogger(false, false),
		Stats:             stats.New(),
		Tracer:            tracer.New(),
		Cache:             cache,
		Scope:             []string{},
	}

	var runErr = runner.Run(&ctx)

	assert.NoError(t, runErr)
}

func Test_BuildSimpleOverrides(t *testing.T) {
	var clearFixture, _, tmpDirAbs = test_helpers.RestoreFixture("simple-overrides")
	defer clearFixture()

	var projectConfigPath, _ = project.FindProjectConfig(tmpDirAbs)
	var cache = cache.New(tmpDirAbs, cache.DefaultCacheLocation)
	cache.Setup()

	var ctx = context.Context{
		Root:              tmpDirAbs,
		Cwd:               tmpDirAbs,
		ProjectConfigPath: projectConfigPath,
		Targets:           []string{"build"},
		Concurrency:       2,
		ChangedFiles:      []string{},
		ChangedOnly:       false,
		Logger:            logger.NewLogger(false, false),
		Stats:             stats.New(),
		Tracer:            tracer.New(),
		Cache:             cache,
		Scope:             []string{},
	}

	var runErr = runner.Run(&ctx)

	assert.NoError(t, runErr)
	assert.FileExists(t, path.Join(tmpDirAbs, "packages", "pkg-a", "dist.js"))
	assert.FileExists(t, path.Join(tmpDirAbs, "packages", "pkg-b", "bundle.js"))
}

// func Test_BuildSimpleOverrides(t *testing.T) {
// 	var tmp_dir = RestoreFixture("simple-overrides")
// 	defer CleanFixture(tmp_dir)

// 	var root_pkg_json, root_config, _ = lib.FindProject(tmp_dir)
// 	var logger = lib.NewLogger(false)
// 	var tracing = lib.NewTracing()
// 	var ctx = lib.NewContext(
// 		tmp_dir,
// 		tmp_dir,
// 		[]string{"build"},
// 		[]string{},
// 		[]string{},
// 		4,
// 		root_pkg_json,
// 		cache.NewCache(tmp_dir),
// 		logger,
// 		tracing,
// 		lib.NewStats(),
// 		root_config,
// 	)
// 	var err = lib.Run(&ctx)

// 	assert.NoError(t, err)
// 	assert.FileExists(t, path.Join(tmp_dir, "packages", "pkg-a", "dist.js"))
// 	assert.FileExists(t, path.Join(tmp_dir, "packages", "pkg-b", "bundle.js"))
// }
