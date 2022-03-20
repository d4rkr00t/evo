package lib_test

import (
	"evo/main/lib"
	"evo/main/lib/cache"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_BuildSimple(t *testing.T) {
	var tmp_dir = RestoreFixture("simple")
	defer CleanFixture(tmp_dir)

	var root_pkg_json, _ = lib.FindRootPackageJson(tmp_dir)
	var logger = lib.NewLogger(false)
	var tracing = lib.NewTracing()
	var ctx = lib.NewContext(
		tmp_dir,
		tmp_dir,
		[]string{"build"},
		[]string{},
		[]string{},
		4,
		root_pkg_json,
		cache.NewCache(tmp_dir),
		logger,
		tracing,
		lib.NewStats(),
		root_pkg_json.GetConfig(),
	)
	var err = lib.Run(ctx)

	assert.NoError(t, err)
	assert.FileExists(t, path.Join(tmp_dir, "packages", "pkg-a", "dist.js"))
	assert.FileExists(t, path.Join(tmp_dir, "packages", "pkg-b", "dist.js"))
}

func Test_BuildSimpleUnknownTarget(t *testing.T) {
	var tmp_dir = RestoreFixture("simple")
	defer CleanFixture(tmp_dir)

	var root_pkg_json, _ = lib.FindRootPackageJson(tmp_dir)
	var logger = lib.NewLogger(false)
	var tracing = lib.NewTracing()
	var ctx = lib.NewContext(
		tmp_dir,
		tmp_dir,
		[]string{"test"},
		[]string{},
		[]string{},
		4,
		root_pkg_json,
		cache.NewCache(tmp_dir),
		logger,
		tracing,
		lib.NewStats(),
		root_pkg_json.GetConfig(),
	)
	var err = lib.Run(ctx)

	assert.NoError(t, err)
}

func Test_BuildSimpleOverrides(t *testing.T) {
	var tmp_dir = RestoreFixture("simple-overrides")
	defer CleanFixture(tmp_dir)

	var root_pkg_json, _ = lib.FindRootPackageJson(tmp_dir)
	var logger = lib.NewLogger(false)
	var tracing = lib.NewTracing()
	var ctx = lib.NewContext(
		tmp_dir,
		tmp_dir,
		[]string{"build"},
		[]string{},
		[]string{},
		4,
		root_pkg_json,
		cache.NewCache(tmp_dir),
		logger,
		tracing,
		lib.NewStats(),
		root_pkg_json.GetConfig(),
	)
	var err = lib.Run(ctx)

	assert.NoError(t, err)
	assert.FileExists(t, path.Join(tmp_dir, "packages", "pkg-a", "dist.js"))
	assert.FileExists(t, path.Join(tmp_dir, "packages", "pkg-b", "bundle.js"))
}
