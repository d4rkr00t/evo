package runner

import (
	"evo/internal/context"
	"evo/internal/project"
	"evo/internal/stats"
)

func BuildDependencyGraph(ctx *context.Context, proj *project.Project) {
	defer ctx.Tracer.Event("dependency graph").Done()
	ctx.Stats.Start("dependency graph", stats.MeasureKindStage)
	var depGraphLg = ctx.Logger.CreateGroup()
	depGraphLg.Debug().Start("Building a dependency graph...")
	proj.BuildDependencyGraph()
	depGraphLg.Debug().EndEmpty(ctx.Stats.Stop("dependency graph"))
}
