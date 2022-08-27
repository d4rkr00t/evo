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
	"path/filepath"

	"github.com/fatih/color"
)

func Hash(ctx *context.Context, labels label.Label) error {
	ctx.Stats.Start("show-hash", stats.MeasureKindStage)
	ctx.Logger.Log()
	ctx.Logger.Badge("root").Log(" " + ctx.Root)
	ctx.Logger.Badge("query").Log("show hash for the label →", color.YellowString(labels.String()))

	if len(labels.Scope) == 0 {
		return errors.New(errors.ErrorEmptyWsName, fmt.Sprintf("Empty workspace. Use `evo show-hash workspace::%s`", labels.Target))
	}

	var proj, err = project.NewProject(ctx.ProjectConfigPath)
	if err != nil {
		return err
	}

	var _, ok = proj.Load(labels.Scope)
	if !ok {
		return errors.New(errors.ErrorWsNotFound, fmt.Sprint("Workspace", labels.Scope, "not found!"))
	}

	err = runner.AugmentDependencies(ctx, &proj)
	if err != nil {
		return err
	}

	proj.ReduceToScope([]string{labels.Scope})
	runner.InvalidateProjects(ctx, &proj)

	var taskGraph = runner.CreateTaskGraph(ctx, &proj)

	taskGraph.Walk(func(task *task_graph.Task) error {
		task.Rehash(taskGraph)
		return nil
	}, ctx.Concurrency)

	var task, taskOk = taskGraph.Load(labels.String())
	if !taskOk {
		return fmt.Errorf("no task named: %s", labels.String())
	}

	var cacheDiag, diagErr = task.RetriveCacheDiagnostics(&ctx.Cache)
	if diagErr == nil {
		if cacheDiag.TaskHash != task.Hash {
			var lgChanged = ctx.Logger.CreateGroup()
			lgChanged.Start(color.HiMagentaString("Task has changed since the last run: %s", formatPairOfHashes(cacheDiag.TaskHash, task.Hash)))

			if cacheDiag.TaskDepsHash != task.DepsHash {
				lgChanged.Log("Task deps:     ", color.YellowString(formatPairOfHashes(cacheDiag.TaskDepsHash, task.DepsHash)))
			}

			if cacheDiag.TaskTarget != task.Target.String() {
				lgChanged.Log("Target:        ", color.YellowString(formatPairOfHashes(cacheDiag.TaskTarget, task.Target.String())))
			}

			if cacheDiag.WsFilesHash != task.Ws.FilesHash {
				lgChanged.Log("Files:         ", color.YellowString(formatPairOfHashes(cacheDiag.WsFilesHash, task.Ws.FilesHash)))
			}

			if cacheDiag.WsExtDepsHash != task.Ws.ExtDepsHash {
				lgChanged.Log("External Deps: ", color.YellowString(formatPairOfHashes(cacheDiag.WsExtDepsHash, task.Ws.ExtDepsHash)))
			}

			if cacheDiag.WsLocalDepsHash != task.Ws.LocalDepsHash {
				lgChanged.Log("Local Deps:    ", color.YellowString(formatPairOfHashes(cacheDiag.WsLocalDepsHash, task.Ws.LocalDepsHash)))
			}

			lgChanged.EndPlain()
		}
	}

	var lg = ctx.Logger.CreateGroup()
	lg.Start(fmt.Sprintf("Hash [%s] consists of:", task.Hash))

	lg.Log("Target:", color.HiBlackString(task.Target.String()))

	lg.Log()
	lg.Log("Files:", color.HiBlackString(task.Ws.FilesHash))
	var files = task.Ws.GetFilesList()
	for _, fileName := range files {
		var filePath, _ = filepath.Rel(task.Ws.Path, fileName)
		lg.Log("–", filePath)
	}

	lg.Log()
	lg.Log("Task Dependencies:", color.HiBlackString(task.DepsHash))
	var taskDeps = task.Deps
	for _, dep := range taskDeps {
		lg.Log(fmt.Sprintf("– %s", dep))
	}

	lg.Log()
	lg.Log("Workspace Dependencies:", color.HiBlackString(fmt.Sprintf("[local:%s] [external:%s]", task.Ws.LocalDepsHash, task.Ws.ExtDepsHash)))
	var wsDeps = task.Ws.Deps
	for _, dep := range wsDeps {
		lg.Log(fmt.Sprintf("– %s:%s [%s] [%s]", dep.Name, dep.Version, dep.Type, dep.Provider))
	}

	lg.End(ctx.Stats.Stop("show-hash"))

	return nil
}

func formatPairOfHashes(h1 string, h2 string) string {
	return fmt.Sprintf("%s → %s", h1[:9], h2[:9])
}
