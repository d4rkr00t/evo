package project

import (
	"evo/internal/context"
	"evo/internal/task_graph"
	"evo/internal/workspace"
	"fmt"
	"strings"
	"sync"
)

func CreateTaskGraphFromProject(ctx *context.Context, proj *Project) task_graph.TaskGraph {
	defer ctx.Tracer.Event("creating task graph").Done()
	var taskGraph = task_graph.New()
	var mutex sync.Mutex

	var addFn = func(task task_graph.Task) {
		mutex.Lock()
		taskGraph.Add(&task)
		for _, depName := range task.Deps {
			taskGraph.Connect(task.Name(), depName)
		}
		mutex.Unlock()
	}

	proj.Walk(func(ws *workspace.Workspace) error {
		defer ctx.Tracer.Event(fmt.Sprintf("creating task graph for %s", ws.Name)).Done()
		for _, targetName := range ctx.Targets {
			createTask(targetName, proj, ws, &addFn)
		}
		return nil
	}, ctx.Concurrency)

	return taskGraph
}

func createTask(targetName string, proj *Project, ws *workspace.Workspace, addFn *func(task task_graph.Task)) {
	var tgt, ok = ws.Targets[targetName]
	if !ok {
		return
	}

	var task = task_graph.NewTask(ws, targetName, &tgt)

	for _, targetDep := range tgt.Deps {
		if isSelfReference(targetDep) {
			createTask(targetDep, proj, ws, addFn)
			task.AddDependency(task_graph.GetTaskName(ws.Name, targetDep))
		} else {
			targetDep = targetDep[1:]

			for _, wsDep := range ws.Deps {
				if wsDep.Type == "external" {
					continue
				}

				var dep, _ = proj.Load(wsDep.Name)
				var _, ok = dep.Targets[targetDep]
				if !ok {
					continue
				}

				task.AddDependency(task_graph.GetTaskName(dep.Name, targetDep))
			}
		}
	}

	(*addFn)(task)
}

func isSelfReference(name string) bool {
	return !strings.HasPrefix(name, "@")
}
