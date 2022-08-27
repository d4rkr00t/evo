package show

import (
	"evo/internal/context"
	"evo/internal/errors"
	"evo/internal/label"
	"evo/internal/project"
	"evo/internal/runner"
	"evo/internal/stats"
	"evo/internal/task_graph"
	"fmt"
	"sync"

	"github.com/fatih/color"
)

func Affected(ctx *context.Context, labels label.Label) error {
	ctx.Stats.Start("show-affected", stats.MeasureKindStage)
	ctx.Logger.Log()
	ctx.Logger.Badge("root").Log(" " + ctx.Root)
	ctx.Logger.Badge("query").Log("show affected targets for the label →", color.YellowString(labels.String()))

	var proj, err = project.NewProject(ctx.ProjectConfigPath)
	if err != nil {
		return err
	}

	if labels.Scope != "" && labels.Scope != "*" {
		var _, ok = proj.Load(labels.Scope)
		if !ok {
			return errors.New(errors.ErrorWsNotFound, fmt.Sprint("Workspace", labels.Scope, "not found!"))
		}
	}

	err = runner.AugmentDependencies(ctx, &proj)
	if err != nil {
		return err
	}

	var scope = label.GetScopeFromLabels(&ctx.Labels)
	if len(scope) > 0 {
		proj.ReduceToScope([]string{labels.Scope})
	}
	runner.InvalidateProjects(ctx, &proj)

	var taskGraph = runner.CreateTaskGraph(ctx, &proj)
	var mu sync.Mutex
	var affectedTargets = []string{}

	taskGraph.Walk(func(task *task_graph.Task) error {
		task.Rehash(taskGraph)
		if task.Invalidate(&ctx.Cache) {
			mu.Lock()
			affectedTargets = append(affectedTargets, task.Name())
			mu.Unlock()
		}
		return nil
	}, ctx.Concurrency)

	var lg = ctx.Logger.CreateGroup()
	lg.Start("Affected targets:")

	if len(affectedTargets) == 0 {
		lg.Log("Nothing is affected")
	} else {
		lg.Log(fmt.Sprintf("%d target(s) affected", len(affectedTargets)))
		lg.Log()
		for _, tgt := range affectedTargets {
			lg.Log("—", tgt)
		}
	}

	lg.End(ctx.Stats.Stop("show-affected"))

	return nil
}
