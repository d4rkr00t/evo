package cache_test

import (
	"evo/internal/cache"
	"evo/internal/test_helpers"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_CacheSetup(t *testing.T) {
	var clearFixture, _, tmpDirAbs = test_helpers.RestoreFixture("project")
	defer clearFixture()

	var cc = cache.New(tmpDirAbs, cache.DefaultCacheLocation)
	cc.Setup()
	assert.DirExists(t, path.Join(tmpDirAbs, cache.DefaultCacheLocation))
}

func Test_CacheHasFalse(t *testing.T) {
	var clearFixture, _, tmpDirAbs = test_helpers.RestoreFixture("project")
	defer clearFixture()

	var cc = cache.New(tmpDirAbs, cache.DefaultCacheLocation)
	cc.Setup()
	assert.False(t, cc.Has("somekey"))
}

func Test_CacheHasTrue(t *testing.T) {
	var clearFixture, _, tmpDirAbs = test_helpers.RestoreFixture("project")
	defer clearFixture()

	var cc = cache.New(tmpDirAbs, cache.DefaultCacheLocation)
	cc.Setup()
	cc.CacheData("somekey", "data")
	assert.True(t, cc.Has("somekey"))
}

func Test_CacheReadData(t *testing.T) {
	var clearFixture, _, tmpDirAbs = test_helpers.RestoreFixture("project")
	defer clearFixture()

	var cc = cache.New(tmpDirAbs, cache.DefaultCacheLocation)
	cc.Setup()
	cc.CacheData("somekey", "data")
	assert.Equal(t, cc.ReadData("somekey"), "data")
}

func Test_CacheReadDataUnknownKey(t *testing.T) {
	var clearFixture, _, tmpDirAbs = test_helpers.RestoreFixture("project")
	defer clearFixture()

	var cc = cache.New(tmpDirAbs, cache.DefaultCacheLocation)
	cc.Setup()
	assert.Equal(t, cc.ReadData("somekey"), "")
}
