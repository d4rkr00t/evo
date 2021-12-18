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
	concurrency   int
	root_pkg_json PackageJson
	cache         cache.Cache
	logger        Logger
	stats         Stats
	config        Config
	tracing       Tracing
	scope         []string
}

func NewContext(
	root string,
	cwd string,
	target []string,
	scope []string,
	concurrency int,
	root_pkg_json PackageJson,
	cache cache.Cache,
	logger Logger,
	tracing Tracing,
	stats Stats,
	config Config,
) Context {
	return Context{
		root, cwd, target, concurrency, root_pkg_json,
		cache, logger, stats, config, tracing, scope,
	}
}

func FindRootPackageJson(cwd string) (PackageJson, error) {
	var pkgjson PackageJson

	for {
		var maybepkgjson_path = path.Join(cwd, "package.json")

		if fileutils.Exist(maybepkgjson_path) {
			var maybepkgjson, _ = NewPackageJson(maybepkgjson_path)
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
