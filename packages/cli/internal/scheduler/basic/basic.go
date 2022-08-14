package basic

import (
	"errors"
	"evo/internal/context"
	"evo/internal/project"
	"evo/internal/stats"
	"evo/internal/task_graph"
	"fmt"
	"sync"
	"sync/atomic"
)

func RunTaskGraph(ctx *context.Context, proj *project.Project, taskGraph *task_graph.TaskGraph) error {
	defer ctx.Tracer.Event("run tasks").Done()
	ctx.Stats.Start("runtasks", stats.MeasureKindStage)
	ctx.Logger.Log()
	ctx.Logger.Log("Running:")
	ctx.Logger.Log()

	ctx.Reporter.Start()

	var taskErrors = sync.Map{}
	var hasErrors int32 = 0
	var taskOutputMutex sync.Mutex

	taskGraph.Walk(func(task *task_graph.Task) error {
		var taskId = task.Name()
		defer ctx.Tracer.Event(fmt.Sprintf("run task %s", taskId)).Done()

		var ws, _ = proj.Load(task.Ws.Name)
		var _, err = RunTask(ctx, taskGraph, task, ws, &taskOutputMutex)

		if err != nil {
			atomic.SwapInt32(&hasErrors, 1)
			taskErrors.Store(taskId, err)
		}

		return nil
	}, ctx.Concurrency)

	ctx.Stats.Stop("runtasks")

	if hasErrors == 1 {
		return errors.New("")
	}

	return nil
}
