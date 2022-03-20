package lib

import (
	"context"
	"errors"
	"fmt"

	"github.com/fatih/color"
	"golang.org/x/sync/semaphore"
)

type TaskResult struct {
	task_id string
	err     error
	out     string
}

func RunTasks(ctx *Context, task_map *TasksMap, wm *WorkspacesMap, lg *LoggerGroup) error {
	var task_errors = []TaskResult{}
	var sem = semaphore.NewWeighted(int64(ctx.concurrency))
	var cc = context.TODO()

	ctx.stats.StartMeasure("runtasks", MEASURE_KIND_STAGE)
	lg.Badge("tasks").Info(fmt.Sprint(task_map.length))
	lg.Log()

	task_map.Walk(func(task_id string) error {
		var task, _ = task_map.Load(task_id)
		task.UpdateStatus(TASK_STATUS_RUNNING)
		task_map.Store(task)

		if err := sem.Acquire(cc, 1); err != nil {
			panic(fmt.Sprintf("Failed to acquire semaphore: %v", err))
		}
		defer sem.Release(1)

		ctx.stats.StartMeasure(task_id, MEASURE_KIND_TASK)

		var trace = ctx.tracing.Event(task_id)
		var out, err = task.Run(ctx, &task)
		trace.Done()
		ctx.stats.StopMeasure(task_id)

		if err == nil {
			lg.Badge(task_id).BadgeColor(task_id).Success("done in " + color.HiBlackString(ctx.stats.GetMeasure(task_id).duration.String()))
			task.UpdateStatus(TASK_STATUS_SUCCESS)
		} else {
			task.UpdateStatus(TASK_STATUS_FAILURE)
			task_errors = append(task_errors, TaskResult{task_id, err, out})
		}

		task_map.Store(task)

		return nil
	})

	lg.Verbose().Log()
	lg.Verbose().Badge("start").Info("   Updating states of workspaces...")
	ctx.stats.StartMeasure("wsstate", MEASURE_KIND_STAGE)
	var trace = ctx.tracing.Event("updating states of workspaces")

	wm.updated.Each(func(ws_name interface{}) bool {
		var ws, _ = wm.Load(ws_name.(string))
		ws.CacheState(&ctx.cache, ws.hash)
		return false
	})

	trace.Done()
	lg.Verbose().Badge("done").Info("    in", ctx.stats.StopMeasure("wsstate").String())
	ctx.stats.StopMeasure("runtasks")

	if !lg.logger.verbose {
		lg.Log()
	}

	if len(task_errors) > 0 {
		lg.Log()
		lg.Log("Errors:")
		lg.Log()
		for _, task_result := range task_errors {
			var task, _ = task_map.Load(task_result.task_id)
			lg.Badge(task_result.task_id).BadgeColor(task.color).Error(task_result.err.Error(), task_result.out)
		}

		return errors.New("")
	}

	return nil
}
