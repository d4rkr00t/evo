package basic

import (
	"errors"
	"evo/internal/clicmd"
	"evo/internal/context"
	"evo/internal/logger"
	"evo/internal/task_graph"
	"evo/internal/workspace"
	"fmt"
	"strings"
	"sync"

	"github.com/fatih/color"
)

func RunTask(ctx *context.Context, taskGraph *task_graph.TaskGraph, task *task_graph.Task, ws *workspace.Workspace, lg *logger.LoggerGroup, taskOutputMutex *sync.Mutex) (string, error) {
	var depsErr = checkStatusesOfTaskDependencies(taskGraph, task)
	var lgCloned = lg.Clone()

	task.UpdateStatus(task_graph.TaskStatsuRunning)
	taskGraph.Store(task)

	if depsErr != nil {
		lgCloned.Badge(task.Name()).Error(color.HiBlackString("← ") + depsErr.Error())
		failTask(ctx, taskGraph, task, "", depsErr)
		return "", depsErr
	}

	if !task.Invalidate(&ctx.Cache) {
		lgCloned.Badge(task.Name()).BadgeColor(task.Color).Success(color.GreenString("cache hit:"), color.HiBlackString(task.WsHash))

		if task.HasOutputs() {
			if task.ShouldRestoreOutputs(&ctx.Cache) {
				task.CleanOutputs()
				task.RestoreOutputs(&ctx.Cache)
				lgCloned.Badge(task.Name()).BadgeColor(task.Color).Debug().Log("outputs don't match, restoring...")
			} else {
				lgCloned.Badge(task.Name()).BadgeColor(task.Color).Debug().Log("outputs match, skip restoring...")
			}
		}

		var taskExitCode, taskOutLogs, taskErrorLogs, taskCacheError = task.GetStatusAndLogs(&ctx.Cache)
		if taskCacheError == nil {
			if len(taskOutLogs) > 0 {
				taskOutputMutex.Lock()
				lgCloned.Badge(task.Name()).BadgeColor(task.Color).Log("replaying output...")
				for _, line := range strings.Split(taskOutLogs, "\n") {
					if taskExitCode == "0" {
						lgCloned.Badge(task.Name()).BadgeColor(task.Color).Log(color.HiBlackString("← "), line)
					} else {
						lgCloned.Badge(task.Name()).BadgeColor(task.Color).Error(color.HiBlackString("← "), line)
					}
				}
				taskOutputMutex.Unlock()
			}
		}

		if taskExitCode == "0" {
			succeedTask(ctx, taskGraph, task, taskOutLogs)
			return taskOutLogs, nil
		} else {
			var err = errors.New(taskErrorLogs)
			failTask(ctx, taskGraph, task, taskOutLogs, err)
			return taskOutLogs, err
		}
	}

	task.CleanOutputs()

	lgCloned.Badge(task.Name()).BadgeColor(task.Color).Log("running →", color.HiBlackString(task.Target.Cmd))

	var cmd = clicmd.NewCmd(
		task.Name(),
		ws.Path,
		task.Target.Cmd,
		func(msg string) {
			lg.Clone().Badge(task.Name()).BadgeColor(task.Color).Log(color.HiBlackString("← ") + msg)
		},
		func(msg string) {
			lg.Clone().Badge(task.Name()).Error(color.HiBlackString("← ") + msg)
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
	taskGraph.Store(task)
	task.Cache(&ctx.Cache, out, "")
}

func failTask(ctx *context.Context, taskGraph *task_graph.TaskGraph, task *task_graph.Task, out string, err error) {
	task.UpdateStatus(task_graph.TaskStatsuFailure)
	taskGraph.Store(task)
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
