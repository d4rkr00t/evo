package basic

import (
	"errors"
	"evo/internal/clicmd"
	"evo/internal/context"
	"evo/internal/stats"
	"evo/internal/task_graph"
	"evo/internal/workspace"
	"fmt"
	"strings"
	"sync"

	"github.com/fatih/color"
)

func RunTask(ctx *context.Context, taskGraph *task_graph.TaskGraph, task *task_graph.Task, ws *workspace.Workspace, taskOutputMutex *sync.Mutex) (string, error) {
	ctx.Stats.Start(task.Name(), stats.MeasureKindTask)
	var depsErr = checkStatusesOfTaskDependencies(taskGraph, task)

	task.UpdateStatus(task_graph.TaskStatsuRunning)
	taskGraph.Store(task)
	ctx.Reporter.UpdateFromTaskGraph(taskGraph)

	if depsErr != nil {
		failTask(ctx, taskGraph, task, "", depsErr)
		return "", depsErr
	}

	if !task.Invalidate(&ctx.Cache, taskGraph) {
		task.RestoredFromCache = task_graph.TaskCacheHit

		if task.HasOutputs() {
			if task.ShouldRestoreOutputs(&ctx.Cache) {
				task.RestoredFromCache = task_graph.TaskCacheHitCopy
				task.CleanOutputs()
				task.RestoreOutputs(&ctx.Cache)
			} else {
				task.RestoredFromCache = task_graph.TaskCacheHitSkip
			}
		}

		var taskExitCode, taskOutLogs, taskErrorLogs, taskCacheError = task.GetStatusAndLogs(&ctx.Cache)
		var err error
		if taskExitCode == "0" {
			succeedTask(ctx, taskGraph, task, taskOutLogs)
		} else {
			err = errors.New(taskErrorLogs)
			failTask(ctx, taskGraph, task, taskOutLogs, err)
		}

		if taskCacheError == nil {
			if len(taskOutLogs) > 0 {
				ctx.Reporter.StreamLog(task, color.YellowString("replaying output…"))
				if taskExitCode == "0" {
					ctx.Reporter.StreamLog(task, strings.Split(taskOutLogs, "\n")...)
				} else {
					ctx.Reporter.StreamError(task, strings.Split(taskOutLogs, "\n")...)
				}
			}
		}

		return taskOutLogs, err
	}

	task.CleanOutputs()

	var cmd = clicmd.NewCmd(
		task.Name(),
		ws.Path,
		task.Target.Cmd,
		func(msg string) {
			ctx.Reporter.StreamLog(task, strings.Split(msg, "\n")...)
		},
		func(msg string) {
			ctx.Reporter.StreamError(task, strings.Split(msg, "\n")...)
		},
	)

	var out, err = cmd.Run()

	if err == nil {
		err = task.ValidateOutputs()
	}

	if err != nil {
		failTask(ctx, taskGraph, task, out, err)
	} else {
		succeedTask(ctx, taskGraph, task, out)
	}

	return out, err
}

func succeedTask(ctx *context.Context, taskGraph *task_graph.TaskGraph, task *task_graph.Task, out string) {
	task.UpdateStatus(task_graph.TaskStatsuSuccess)
	task.Duration = ctx.Stats.Stop(task.Name())
	task.Output = out
	taskGraph.Store(task)

	ctx.Reporter.SuccessTask(task)
	ctx.Reporter.UpdateFromTaskGraph(taskGraph)

	task.Cache(&ctx.Cache, out, "")
}

func failTask(ctx *context.Context, taskGraph *task_graph.TaskGraph, task *task_graph.Task, out string, err error) {
	task.UpdateStatus(task_graph.TaskStatsuFailure)
	task.Duration = ctx.Stats.Stop(task.Name())
	task.Output = out
	task.Error = err
	taskGraph.Store(task)

	ctx.Reporter.FailTask(task)
	ctx.Reporter.UpdateFromTaskGraph(taskGraph)

	task.Cache(&ctx.Cache, out, err.Error())
}

func checkStatusesOfTaskDependencies(taskGraph *task_graph.TaskGraph, task *task_graph.Task) error {
	for _, depName := range task.Deps {
		var depTask, _ = taskGraph.Load(depName)
		if depTask.Status == task_graph.TaskStatsuFailure {
			return fmt.Errorf(fmt.Sprintf("cannot continue, dependency \"%s\" has failed", color.CyanString(depName)))
		}
	}

	return nil
}
