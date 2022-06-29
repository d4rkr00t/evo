package runner_test

import (
	"evo/internal/cache"
	"evo/internal/context"
	"evo/internal/logger"
	"evo/internal/project"
	"evo/internal/reporter"
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

	var logger = logger.NewLogger(false, false)
	var rr = reporter.New(logger)
	var ctx = context.Context{
		Root:              tmpDirAbs,
		Cwd:               tmpDirAbs,
		ProjectConfigPath: projectConfigPath,
		Targets:           []string{"build"},
		Concurrency:       2,
		ChangedFiles:      []string{},
		ChangedOnly:       false,
		Logger:            logger,
		Stats:             stats.New(),
		Tracer:            tracer.New(),
		Reporter:          &rr,
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

	var logger = logger.NewLogger(false, false)
	var rr = reporter.New(logger)
	var ctx = context.Context{
		Root:              tmpDirAbs,
		Cwd:               tmpDirAbs,
		ProjectConfigPath: projectConfigPath,
		Targets:           []string{"test"},
		Concurrency:       2,
		ChangedFiles:      []string{},
		ChangedOnly:       false,
		Logger:            logger,
		Reporter:          &rr,
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

	var logger = logger.NewLogger(false, false)
	var rr = reporter.New(logger)
	var ctx = context.Context{
		Root:              tmpDirAbs,
		Cwd:               tmpDirAbs,
		ProjectConfigPath: projectConfigPath,
		Targets:           []string{"build"},
		Concurrency:       2,
		ChangedFiles:      []string{},
		ChangedOnly:       false,
		Logger:            logger,
		Reporter:          &rr,
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
