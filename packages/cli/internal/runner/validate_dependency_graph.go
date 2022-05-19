package runner

import (
	"evo/internal/context"
	"evo/internal/project"
	"evo/internal/stats"
)

func ValidateDependencyGraph(ctx *context.Context, proj *project.Project) error {
	defer ctx.Tracer.Event("validating dependency graph").Done()
	ctx.Stats.Start("validating dependency graph", stats.MeasureKindStage)

	var depGraphValidationLg = ctx.Logger.CreateGroup()
	depGraphValidationLg.Debug().Start("Validating a dependency graph...")

	var depGraphErr = proj.Validate()

	depGraphValidationLg.Debug().EndEmpty(ctx.Stats.Stop("validating dependency graph"))

	return depGraphErr
}
