package context

import (
	"evo/internal/cache"
	"evo/internal/logger"
	"evo/internal/stats"
	"evo/internal/tracer"
)

type Context struct {
	Root              string
	Cwd               string
	ProjectConfigPath string
	Targets           []string
	Concurrency       int
	Logger            logger.Logger
	Stats             stats.Stats
	Tracer            tracer.Tracer
	Cache             cache.Cache
	Scope             []string
	ChangedFiles      []string
	ChangedOnly       bool
}
