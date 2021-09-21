package lib

import (
	"errors"
	"evo/main/lib/cache"
	"evo/main/lib/fileutils"
	"path"
)

type Context struct {
	root          string
	cwd           string
	target        []string
	root_pkg_json PackageJson
	cache         cache.Cache
	logger        Logger
	stats         Stats
	config        Config
}

func NewContext(
	root string,
	cwd string,
	target []string,
	root_pkg_json PackageJson,
	cache cache.Cache,
	logger Logger,
	stats Stats,
	config Config,
) Context {
	return Context{
		root, cwd, target, root_pkg_json,
		cache, logger, stats, config,
	}
}

func FindRootPackageJson(cwd string) (PackageJson, error) {
	var pkgjson PackageJson

	for {
		var maybepkgjson_path = path.Join(cwd, "package.json")

		if fileutils.Exist(maybepkgjson_path) {
			var maybepkgjson = NewPackageJson(maybepkgjson_path)
			if len(maybepkgjson.Evo.Workspaces) > 0 {
				return maybepkgjson, nil
			}
		}

		if cwd == path.Dir(cwd) {
			break
		}

		cwd = path.Dir(cwd)
	}

	return pkgjson, errors.New("not in evo project")
}
