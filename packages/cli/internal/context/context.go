package context

import (
	"evo/internal/cache"
	"evo/internal/label"
	"evo/internal/logger"
	"evo/internal/reporter"
	"evo/internal/stats"
	"evo/internal/tracer"
)

type Context struct {
	Root              string
	Cwd               string
	ProjectConfigPath string
	Labels            []label.Label
	Concurrency       int
	Logger            logger.Logger
	Reporter          *reporter.Reporter
	Stats             stats.Stats
	Tracer            tracer.Tracer
	Cache             cache.Cache
	ChangedFiles      []string
	ChangedOnly       bool
}
