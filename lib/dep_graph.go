package lib

import (
	"sync"

	"github.com/pyr-sh/dag"
)

func NewDAGFromWorkspaces(workspaces *sync.Map) dag.AcyclicGraph {
	var graph dag.AcyclicGraph

	workspaces.Range(func(_ interface{}, value interface{}) bool {
		var ws = value.(Workspace)

		graph.Add(ws.Name)

		for name := range ws.Deps {
			if _, ok := workspaces.Load(name); ok {
				graph.Connect(dag.BasicEdge(ws.Name, name))
			}
		}

		return true
	})

	return graph
}
