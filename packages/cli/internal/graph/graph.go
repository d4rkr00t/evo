package graph

import (
	"evo/internal/context"
	"evo/internal/label"
	"evo/internal/project"
	"evo/internal/runner"
	"fmt"
	"strings"
)

func Graph(ctx *context.Context) error {
	var proj, err = project.NewProject(ctx.ProjectConfigPath)
	if err != nil {
		return err
	}

	err = runner.AugmentDependencies(ctx, &proj)
	if err != nil {
		return err
	}

	var scope = label.GetScopeFromLabels(&ctx.Labels)
	if len(scope) > 0 {
		proj.ReduceToScope(scope)
	}

	var taskGraph = runner.CreateTaskGraph(ctx, &proj)
	err = runner.ValidateTaskGraph(ctx, taskGraph)
	if err != nil {
		return err
	}

	var visited = map[string]bool{}
	var shapes = []string{}
	var nodes = []string{}
	for _, wsName := range proj.WorkspacesNames {
		var from = fmt.Sprintf("\"%s\"", wsName)
		if _, ok := visited[wsName]; !ok {
			visited[wsName] = true
			shapes = append(shapes, fmt.Sprintf("%s [shape=doubleoctagon]", from))
		}

		var ws, _ = proj.Load(wsName)

		for _, dep := range ws.Deps {
			var to = fmt.Sprintf("\"%s\"", dep.Name)

			if _, ok := visited[dep.Name]; !ok {
				visited[dep.Name] = true
				if dep.Type == "external" {
					nodes = append(nodes, fmt.Sprintf("%s -> %s", from, to))
					shapes = append(shapes, fmt.Sprintf("%s [fillcolor=lightgoldenrodyellow, style=\"filled\"]", to))
				} else {
					shapes = append(shapes, fmt.Sprintf("%s [shape=doubleoctagon]", to))
				}
			}
		}
	}

	for _, taskName := range taskGraph.TasksNamesList {
		var task, _ = taskGraph.Load(taskName)
		var fromWs = fmt.Sprintf("\"%s\"", task.WsName)
		var from = fmt.Sprintf("\"%s\"", taskName)
		if _, ok := visited[taskName]; !ok {
			visited[taskName] = true
			shapes = append(shapes, fmt.Sprintf("%s [shape=box, fillcolor=gray96, style=\"filled\"]", from))
			nodes = append(nodes, fmt.Sprintf("%s -> %s", fromWs, from))
		}

		for _, dep := range task.Deps {
			var to = fmt.Sprintf("\"%s\"", dep)
			nodes = append(nodes, fmt.Sprintf("%s -> %s", from, to))

			if _, ok := visited[dep]; !ok {
				visited[dep] = true
				shapes = append(shapes, fmt.Sprintf("%s [shape=box, fillcolor=gray96, style=\"filled\"]", to))
			}
		}
	}

	fmt.Println("digraph G {")
	fmt.Println("{")
	fmt.Print(strings.Join(shapes, "\n"))
	fmt.Println("}")
	fmt.Print(strings.Join(nodes, "\n"))
	fmt.Println("}")

	return nil
}
