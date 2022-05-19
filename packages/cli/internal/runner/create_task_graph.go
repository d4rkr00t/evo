package runner

import (
	"evo/internal/context"
	"evo/internal/project"
	"evo/internal/stats"
	"evo/internal/task_graph"
)

func CreateTaskGraph(ctx *context.Context, proj *project.Project) task_graph.TaskGraph {
	ctx.Stats.Start("task graph", stats.MeasureKindStage)
	var taskGraphLg = ctx.Logger.CreateGroup()
	taskGraphLg.Debug().Start("Building a task graph...")
	var taskGraph = project.CreateTaskGraphFromProject(ctx, proj)
	taskGraphLg.Debug().EndEmpty(ctx.Stats.Stop("task graph"))
	return taskGraph
}
