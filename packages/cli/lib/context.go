package lib

import (
	"evo/main/lib/cache"
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
	changed_files []string
}

func NewContext(
	root string,
	cwd string,
	target []string,
	scope []string,
	changed_files []string,
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
		cache, logger, stats, config, tracing, scope, changed_files,
	}
}
