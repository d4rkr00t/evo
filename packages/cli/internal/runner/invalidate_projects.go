package runner

import (
	"evo/internal/context"
	"evo/internal/project"
	"evo/internal/stats"
)

func InvalidateProjects(ctx *context.Context, proj *project.Project) {
	defer ctx.Tracer.Event("invalidating projects").Done()
	ctx.Stats.Start("invalidaing projects", stats.MeasureKindStage)
	var depGraphLg = ctx.Logger.CreateGroup()
	depGraphLg.Debug().Start("Invalidating projects...")
	proj.RehashAllWorkspaces(ctx)
	depGraphLg.Debug().EndEmpty(ctx.Stats.Stop("invalidaing projects"))
}
