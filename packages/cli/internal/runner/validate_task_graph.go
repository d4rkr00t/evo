package runner

import (
	"evo/internal/context"
	"evo/internal/stats"
	"evo/internal/task_graph"
)

func ValidateTaskGraph(ctx *context.Context, taskGraph *task_graph.TaskGraph) error {
	defer ctx.Tracer.Event("validating task graph").Done()
	ctx.Stats.Start("validating task graph", stats.MeasureKindStage)

	var taskGraphValidationLg = ctx.Logger.CreateGroup()
	taskGraphValidationLg.Debug().Start("Validating a task graph...")

	var taskGraphErr = taskGraph.Validate()

	taskGraphValidationLg.Debug().EndEmpty(ctx.Stats.Stop("validating task graph"))

	return taskGraphErr
}
