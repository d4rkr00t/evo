package project

import (
	"evo/internal/context"
	"evo/internal/task_graph"
	"evo/internal/workspace"
	"fmt"
	"strings"
)

func CreateTaskGraphFromProject(ctx *context.Context, proj *Project) *task_graph.TaskGraph {
	defer ctx.Tracer.Event("creating task graph").Done()
	var taskGraph = task_graph.New()

	proj.Range(func(ws *workspace.Workspace) bool {
		defer ctx.Tracer.Event(fmt.Sprintf("creating task graph for %s", ws.Name)).Done()
		for _, label := range ctx.Labels {
			if label.Scope == "*" || label.Scope == ws.Name {
				if label.Target == "*" {
					for targetName := range ws.Targets {
						createTask(proj, &taskGraph, targetName, ws, true)
					}
				} else {
					createTask(proj, &taskGraph, label.Target, ws, true)
				}
			}
		}
		return true
	})

	return &taskGraph
}

func createTask(proj *Project, taskGraph *task_graph.TaskGraph, targetName string, ws *workspace.Workspace, topLevel bool) {
	var tgt, ok = ws.Targets[targetName]
	if !ok {
		return
	}

	var task = task_graph.NewTask(ws, targetName, &tgt, topLevel)

	for _, targetDep := range tgt.Deps {
		if isSelfReference(targetDep) {
			createTask(proj, taskGraph, targetDep, ws, false)
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

				createTask(proj, taskGraph, targetDep, dep, false)
				task.AddDependency(task_graph.GetTaskName(dep.Name, targetDep))
			}
		}
	}

	taskGraph.Add(&task)
}

func isSelfReference(name string) bool {
	return !strings.HasPrefix(name, "@")
}
