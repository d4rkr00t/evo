package lib

import "evo/main/lib/cache"

type Context struct {
	root          string
	cwd           string
	target        string
	root_pkg_json PackageJson
	cache         cache.Cache
	logger        Logger
	stats         Stats
	config        Config
}

func NewContext(
	root string,
	cwd string,
	target string,
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
