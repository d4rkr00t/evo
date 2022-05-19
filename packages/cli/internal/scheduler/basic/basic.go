package basic

import (
	"errors"
	"evo/internal/context"
	"evo/internal/project"
	"evo/internal/stats"
	"evo/internal/task_graph"
	"fmt"
	"strings"
	"sync"

	"github.com/fatih/color"
)

type TaskResult struct {
	taskID  string
	taskErr error
}

func RunTaskGraph(ctx *context.Context, proj *project.Project, taskGraph *task_graph.TaskGraph) error {
	defer ctx.Tracer.Event("run tasks").Done()
	var lg = ctx.Logger.CreateGroup()
	ctx.Stats.Start("runtasks", stats.MeasureKindStage)
	lg.Start(fmt.Sprintf("Running targets → %s", color.CyanString(strings.Join(ctx.Targets, ", "))))

	var taskErrors = sync.Map{}
	var taskOutputMutex sync.Mutex

	taskGraph.Walk(func(task *task_graph.Task) error {
		var taskId = task.Name()
		defer ctx.Tracer.Event(fmt.Sprintf("run task %s", taskId)).Done()
		ctx.Stats.Start(taskId, stats.MeasureKindTask)

		var ws, _ = proj.Load(task.WsName)
		var _, err = RunTask(ctx, taskGraph, task, ws, &lg, &taskOutputMutex)

		if err != nil {
			taskErrors.Store(taskId, err)
		}

		ctx.Stats.Stop(taskId)
		return nil
	}, ctx.Concurrency)

	var taskErrorsList = []TaskResult{}

	taskErrors.Range(func(key, value any) bool {
		var taskID = key.(string)
		var err = value.(error)
		taskErrorsList = append(taskErrorsList, TaskResult{taskID: taskID, taskErr: err})
		return true
	})

	if len(taskErrorsList) > 0 {
		lg.Log()
		lg.Log("Errors:")
		lg.Log()
		for _, taskResult := range taskErrorsList {
			lg.Badge(taskResult.taskID).Error(taskResult.taskErr.Error())
		}

		lg.Fail(ctx.Stats.Stop("runtasks"))
		return errors.New("")
	}

	lg.End(ctx.Stats.Stop("runtasks"))
	return nil
}
